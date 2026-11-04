package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

const cachedLinks = `{
	%s
	"folders": {"plant": {"links": [{"title": "Runbook", "url": "https://example.com/runbook"}]}},
	"queryLinks": [
		{"id": "l1", "name": "plant.temp", "query": {"expr": "temp", "datasource": {"uid": "prom", "type": "prometheus"}}},
		{"id": "nocache", "name": "plant.live", "query": {"expr": "live", "datasource": {"uid": "prom", "type": "prometheus"}},
		 "metadata": {"cacheTtl": 0}}
	]
}`

// queryAt runs one query over [from, to]
func queryAt(t *testing.T, ds *Datasource, refID, linkID string, from, to time.Time, headers map[string]string) backend.DataResponse {
	t.Helper()
	res, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Headers: headers,
		Queries: []backend.DataQuery{{
			RefID:         refID,
			JSON:          []byte(fmt.Sprintf(`{"linkId": %q}`, linkID)),
			Interval:      time.Minute,
			MaxDataPoints: 100,
			TimeRange:     backend.TimeRange{From: from, To: to},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := res.Responses[refID]
	if r.Error != nil {
		t.Fatal(r.Error)
	}
	return r
}

func cacheCount(result string) float64 {
	return testutil.ToFloat64(cacheRequests.WithLabelValues(result))
}

func TestQueryCache(t *testing.T) {
	mr := miniredis.RunT(t)
	for name, cacheConfig := range map[string]string{
		"memory": `"cache": {"defaultTtl": 60},`,
		"redis":  fmt.Sprintf(`"cache": {"defaultTtl": 60, "redisAddress": %q},`, mr.Addr()),
	} {
		t.Run(name, func(t *testing.T) {
			mr.FlushAll()
			g := newFakeGrafana(t)
			ds := newTestDatasource(t, g.URL, testToken, fmt.Sprintf(cachedLinks, cacheConfig))
			now := time.Date(2026, 11, 4, 12, 0, 30, 0, time.UTC)
			hour := now.Add(-time.Hour)

			first := queryAt(t, ds, "A", "l1", hour, now, nil)
			// Another panel, a few seconds later, in the same minute: answered from the cache
			hits := cacheCount("hit")
			second := queryAt(t, ds, "B", "l1", hour.Add(5*time.Second), now.Add(5*time.Second), nil)
			if n := g.queryCount(); n != 1 {
				t.Fatalf("want 1 query to the source, got %d", n)
			}
			if cacheCount("hit") != hits+1 {
				t.Error("the second query wasn't counted as a hit")
			}
			if second.Frames[0].RefID != "B" || first.Frames[0].RefID != "A" {
				t.Errorf("frames keep their own panel's refId: %s, %s", first.Frames[0].RefID, second.Frames[0].RefID)
			}
			// Decorated after the cache, once: metadata changes apply without waiting for it
			if links := second.Frames[0].Fields[1].Config.Links; len(links) != 1 {
				t.Errorf("want the runbook link once on a cached answer, got %v", links)
			}

			// The range sent to the source is rounded to the minute
			if g.requests[0]["from"] != fmt.Sprint(hour.Truncate(time.Minute).UnixMilli()) {
				t.Errorf("range not aligned: from %v", g.requests[0]["from"])
			}

			// The next minute is a new entry
			queryAt(t, ds, "A", "l1", hour.Add(time.Minute), now.Add(time.Minute), nil)
			if n := g.queryCount(); n != 2 {
				t.Errorf("want a new query for the next minute, got %d queries", n)
			}

			// Alert rules always go to the source, and say so to it
			queryAt(t, ds, "A", "l1", hour, now, map[string]string{backend.FromAlertHeaderName: "true"})
			if n := g.queryCount(); n != 3 {
				t.Errorf("an alert query was answered from the cache")
			}
			if h := g.headers[len(g.headers)-1].Get(backend.FromAlertHeaderName); h != "true" {
				t.Errorf("FromAlert not forwarded to the source: %q", h)
			}

			// A link that turns the cache off
			queryAt(t, ds, "A", "nocache", hour, now, nil)
			queryAt(t, ds, "A", "nocache", hour, now, nil)
			if n := g.queryCount(); n != 5 {
				t.Errorf("a link with cacheTtl 0 was cached: %d queries", n)
			}
		})
	}
}

func TestQueryCacheSharesQueriesInFlight(t *testing.T) {
	g := newFakeGrafana(t)
	g.delay = 200 * time.Millisecond
	ds := newTestDatasource(t, g.URL, testToken, fmt.Sprintf(cachedLinks, ""))
	now := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			queryAt(t, ds, fmt.Sprintf("P%d", i), "l1", now.Add(-time.Hour), now, nil)
		}(i)
	}
	wg.Wait()
	if n := g.queryCount(); n != 1 {
		t.Errorf("5 identical queries at once should make 1 trip to the source, made %d", n)
	}
}

func TestQueryCacheFailureFallsBackToTheSource(t *testing.T) {
	mr := miniredis.RunT(t)
	addr := mr.Addr()
	mr.Close() // Redis is down

	g := newFakeGrafana(t)
	ds := newTestDatasource(t, g.URL, testToken, fmt.Sprintf(cachedLinks, fmt.Sprintf(`"cache": {"redisAddress": %q},`, addr)))
	errs := cacheCount("error")
	now := time.Now()
	queryAt(t, ds, "A", "l1", now.Add(-time.Hour), now, nil)
	if g.queryCount() != 1 || cacheCount("error") != errs+1 {
		t.Error("a query with Redis down should still be answered, and the failure counted")
	}

	res, _ := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if res.Status != backend.HealthStatusError || !strings.Contains(res.Message, "Redis at "+addr) {
		t.Errorf("the health check should report Redis being down, got %q", res.Message)
	}
}

func TestAlignRange(t *testing.T) {
	at := func(s string) time.Time { v, _ := time.Parse(time.RFC3339, s); return v }
	long := backend.TimeRange{From: at("2026-11-04T11:00:42Z"), To: at("2026-11-04T12:00:42Z")}
	if got := alignRange(long, time.Minute); got.From != at("2026-11-04T11:00:00Z") || got.To != at("2026-11-04T12:00:00Z") {
		t.Errorf("long range: got %v", got)
	}
	short := backend.TimeRange{From: at("2026-11-04T11:55:42Z"), To: at("2026-11-04T12:00:42Z")}
	if got := alignRange(short, time.Minute); got != short {
		t.Errorf("a range under ten ttls shouldn't move: got %v", got)
	}
}

func TestMemoryStoreStaysWithinItsLimit(t *testing.T) {
	m := &memoryStore{entries: map[string]memoryEntry{}}
	big := make([]byte, memoryLimit/8)
	for i := 0; i < 20; i++ {
		_ = m.Set(context.Background(), fmt.Sprint(i), big, time.Duration(i+1)*time.Minute)
	}
	if m.size > memoryLimit {
		t.Errorf("size %d is over the limit", m.size)
	}
	if _, ok, _ := m.Get(context.Background(), "19"); !ok {
		t.Error("the latest entry should be kept")
	}
	if _, ok, _ := m.Get(context.Background(), "0"); ok {
		t.Error("the entry expiring soonest should have been evicted")
	}
}
