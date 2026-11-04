package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"github.com/srclosson/object-datasource/pkg/models"
)

// keyPrefix namespaces the cache's keys in a shared Redis, and changes when the format does
const keyPrefix = "grafana-object-datasource:v1:"

// memoryLimit bounds the in-memory cache
const memoryLimit = 64 << 20

var cacheRequests = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "grafana_object_datasource",
	Name:      "cache_requests_total",
	Help:      "Queries by how the cache answered: hit, miss, shared (joined an identical query in flight), bypass (not cached), or error (the cache failed and the query went to the source).",
}, []string{"result"})

// store keeps cached responses: Redis, shared between plugin processes, or memory
type store interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Ping(ctx context.Context) error
	Close() error
	Describe() string
}

func newStore(s models.CacheSettings) store {
	if s.RedisAddress == "" {
		return &memoryStore{entries: map[string]memoryEntry{}}
	}
	return &redisStore{
		address: s.RedisAddress,
		client:  redis.NewClient(&redis.Options{Addr: s.RedisAddress, Password: s.RedisPassword, DB: s.RedisDB}),
	}
}

type redisStore struct {
	address string
	client  *redis.Client
}

func (r *redisStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	b, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	return b, err == nil, err
}

func (r *redisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisStore) Ping(ctx context.Context) error { return r.client.Ping(ctx).Err() }
func (r *redisStore) Close() error                   { return r.client.Close() }
func (r *redisStore) Describe() string               { return "Redis at " + r.address }

type memoryStore struct {
	mu      sync.Mutex
	entries map[string]memoryEntry
	size    int
}

type memoryEntry struct {
	value   []byte
	expires time.Time
}

func (m *memoryStore) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.entries[key]
	if !ok || time.Now().After(e.expires) {
		return nil, false, nil
	}
	return e.value, true, nil
}

func (m *memoryStore) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if len(value) > memoryLimit/4 {
		return nil // too big to be worth holding
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	if old, ok := m.entries[key]; ok {
		m.size -= len(old.value)
		delete(m.entries, key)
	}
	// Make room: expired entries first, then whichever expire soonest
	for m.size+len(value) > memoryLimit && len(m.entries) > 0 {
		var victim string
		var soonest time.Time
		for k, e := range m.entries {
			if e.expires.Before(now) {
				victim = k
				break
			}
			if victim == "" || e.expires.Before(soonest) {
				victim, soonest = k, e.expires
			}
		}
		m.size -= len(m.entries[victim].value)
		delete(m.entries, victim)
	}
	m.entries[key] = memoryEntry{value: value, expires: now.Add(ttl)}
	m.size += len(value)
	return nil
}

func (m *memoryStore) Ping(context.Context) error { return nil }
func (m *memoryStore) Close() error               { return nil }
func (m *memoryStore) Describe() string           { return "in memory" }

// queryCache answers repeated queries from the store, and lets identical queries that arrive
// together share one trip to the source
type queryCache struct {
	store store
	group singleflight.Group
}

// alignRange rounds a time range down to multiples of ttl, so that refreshes of a moving range
// like "the last hour" within the same ttl share a cache entry. Ranges shorter than ten ttls
// aren't moved: rounding would cut off too much of them.
func alignRange(tr backend.TimeRange, ttl time.Duration) backend.TimeRange {
	if ttl <= 0 || tr.To.Sub(tr.From) < 10*ttl {
		return tr
	}
	return backend.TimeRange{From: tr.From.Truncate(ttl), To: tr.To.Truncate(ttl)}
}

func cacheKey(orgID int64, model map[string]any, tr backend.TimeRange) string {
	b, _ := json.Marshal([]any{orgID, model, tr.From.UnixMilli(), tr.To.UnixMilli()})
	sum := sha256.Sum256(b)
	return keyPrefix + hex.EncodeToString(sum[:])
}

// query answers from the cache when it can. ttl of zero, or a query from an alert rule, goes
// straight to the source. Only successful responses are cached, and if the cache fails, the
// query goes to the source anyway.
func (c *queryCache) query(ctx context.Context, orgID int64, model map[string]any, tr backend.TimeRange, ttl time.Duration, fromAlert bool,
	fetch func(backend.TimeRange) (backend.DataResponse, error)) (backend.DataResponse, error) {
	if ttl <= 0 || fromAlert {
		cacheRequests.WithLabelValues("bypass").Inc()
		return fetch(tr)
	}
	tr = alignRange(tr, ttl)
	key := cacheKey(orgID, model, tr)

	if b, ok, err := c.store.Get(ctx, key); err != nil {
		backend.Logger.Warn("Query cache read failed; querying the source", "cache", c.store.Describe(), "error", err)
		cacheRequests.WithLabelValues("error").Inc()
		return fetch(tr)
	} else if ok {
		if frames, err := decodeFrames(b); err == nil {
			cacheRequests.WithLabelValues("hit").Inc()
			return backend.DataResponse{Frames: frames}, nil
		}
	}

	type result struct {
		res backend.DataResponse
		err error
	}
	v, _, shared := c.group.Do(key, func() (any, error) {
		res, err := fetch(tr)
		if err == nil && res.Error == nil {
			if b, encErr := encodeFrames(res.Frames); encErr == nil {
				// Not the request's context: the entry should be written even if this caller has gone
				setCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if setErr := c.store.Set(setCtx, key, b, ttl); setErr != nil {
					backend.Logger.Warn("Query cache write failed", "cache", c.store.Describe(), "error", setErr)
				}
			}
		}
		return result{res, err}, nil
	})
	if shared {
		cacheRequests.WithLabelValues("shared").Inc()
	} else {
		cacheRequests.WithLabelValues("miss").Inc()
	}
	r := v.(result)
	if shared {
		// Callers decorate the frames they get, so each needs its own copy
		r.res.Frames = copyFrames(r.res.Frames)
	}
	return r.res, r.err
}

func encodeFrames(frames data.Frames) ([]byte, error) {
	b, err := frames.MarshalArrow()
	if err != nil {
		return nil, err
	}
	return json.Marshal(b)
}

func decodeFrames(b []byte) (data.Frames, error) {
	var encoded [][]byte
	if err := json.Unmarshal(b, &encoded); err != nil {
		return nil, err
	}
	return data.UnmarshalArrowFrames(encoded)
}

func copyFrames(frames data.Frames) data.Frames {
	b, err := encodeFrames(frames)
	if err != nil {
		return frames
	}
	out, err := decodeFrames(b)
	if err != nil {
		return frames
	}
	return out
}
