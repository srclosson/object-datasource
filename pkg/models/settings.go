package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// DataSourceRef identifies the datasource a query link runs against
type DataSourceRef struct {
	UID  string `json:"uid"`
	Type string `json:"type,omitempty"`
}

// QueryLink is a query written once by someone who knows the target datasource, and given a
// place in the tree by its dot-separated name
type QueryLink struct {
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name,omitempty"`
	UID      string          `json:"uid,omitempty"`
	Query    json.RawMessage `json:"query,omitempty"`
	Metadata Metadata        `json:"metadata,omitempty"`
}

// Datasource is the target datasource
func (l QueryLink) Datasource() DataSourceRef {
	return datasourceOf(l.Query, l.UID)
}

// datasourceOf is the datasource a query names, falling back to uid
func datasourceOf(query json.RawMessage, uid string) DataSourceRef {
	var q struct {
		Datasource *DataSourceRef `json:"datasource"`
	}
	_ = json.Unmarshal(query, &q)
	if q.Datasource != nil && q.Datasource.UID != "" {
		return *q.Datasource
	}
	return DataSourceRef{UID: uid}
}

type jsonData struct {
	QueryLinks []QueryLink         `json:"queryLinks,omitempty"`
	Folders    map[string]Metadata `json:"folders,omitempty"`
	Cache      CacheSettings       `json:"cache,omitempty"`
}

// CacheSettings configure the query cache
type CacheSettings struct {
	// DefaultTTL is how long results are reused, in seconds, unless a folder or link says
	// otherwise. Zero uses DefaultCacheTTL; a negative value turns the cache off.
	DefaultTTL int `json:"defaultTtl,omitempty"`
	// RedisAddress is host:port of a Redis server to share the cache through. Without one,
	// each plugin process caches in memory.
	RedisAddress string `json:"redisAddress,omitempty"`
	RedisDB      int    `json:"redisDb,omitempty"`
	// RedisPassword comes from the secure settings
	RedisPassword string `json:"-"`
}

// DefaultCacheTTL is how long results are reused, in seconds, when the settings don't say
const DefaultCacheTTL = 30

// Settings are the datasource instance's settings
type Settings struct {
	// GrafanaURL overrides Grafana's root_url for calls to its API, for when root_url isn't
	// reachable from the plugin (for example inside a container)
	GrafanaURL string
	// ServiceAccountToken is a token pasted into the datasource settings. Used when Grafana
	// hasn't given the plugin a managed service account.
	ServiceAccountToken string
	QueryLinks          []QueryLink
	// Folders is the metadata of folders in the tree, by path
	Folders map[string]Metadata
	Cache   CacheSettings
}

// LoadSettings reads Settings from the datasource configuration
func LoadSettings(config backend.DataSourceInstanceSettings) (*Settings, error) {
	var jd jsonData
	if len(config.JSONData) > 0 {
		if err := json.Unmarshal(config.JSONData, &jd); err != nil {
			return nil, fmt.Errorf("could not read the datasource settings: %w", err)
		}
	}
	settings := &Settings{
		GrafanaURL:          strings.TrimRight(strings.TrimSpace(config.URL), "/"),
		ServiceAccountToken: strings.TrimSpace(config.DecryptedSecureJSONData["serviceAccountToken"]),
		QueryLinks:          jd.QueryLinks,
		Folders:             jd.Folders,
		Cache:               jd.Cache,
	}
	settings.Cache.RedisAddress = strings.TrimSpace(settings.Cache.RedisAddress)
	settings.Cache.RedisPassword = config.DecryptedSecureJSONData["redisPassword"]
	return settings, nil
}

// CacheTTL is how long a link's results are reused: its own or its folders' setting, or the
// datasource's default. Zero means not cached.
func (s *Settings) CacheTTL(md Metadata) time.Duration {
	ttl := s.Cache.DefaultTTL
	if ttl == 0 {
		ttl = DefaultCacheTTL
	}
	if md.CacheTTL != nil {
		ttl = *md.CacheTTL
	}
	if ttl < 0 {
		return 0
	}
	return time.Duration(ttl) * time.Second
}

// FindLink returns the query link with this id, or failing that, this name
func (s *Settings) FindLink(id, name string) (QueryLink, bool) {
	for _, l := range s.QueryLinks {
		if id != "" && l.ID == id {
			return l, true
		}
	}
	for _, l := range s.QueryLinks {
		if name != "" && l.Name == name {
			return l, true
		}
	}
	return QueryLink{}, false
}
