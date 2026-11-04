package models

import (
	"encoding/json"
	"fmt"
	"strings"

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
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	UID   string          `json:"uid,omitempty"`
	Query json.RawMessage `json:"query,omitempty"`
}

// Datasource is the target datasource, taken from the link's query and falling back to its uid
func (l QueryLink) Datasource() DataSourceRef {
	var q struct {
		Datasource *DataSourceRef `json:"datasource"`
	}
	_ = json.Unmarshal(l.Query, &q)
	if q.Datasource != nil && q.Datasource.UID != "" {
		return *q.Datasource
	}
	return DataSourceRef{UID: l.UID}
}

type jsonData struct {
	QueryLinks []QueryLink `json:"queryLinks,omitempty"`
}

// Settings are the datasource instance's settings
type Settings struct {
	// GrafanaURL overrides Grafana's root_url for calls to its API, for when root_url isn't
	// reachable from the plugin (for example inside a container)
	GrafanaURL string
	// ServiceAccountToken is a token pasted into the datasource settings. Used when Grafana
	// hasn't given the plugin a managed service account.
	ServiceAccountToken string
	QueryLinks          []QueryLink
}

// LoadSettings reads Settings from the datasource configuration
func LoadSettings(config backend.DataSourceInstanceSettings) (*Settings, error) {
	var jd jsonData
	if len(config.JSONData) > 0 {
		if err := json.Unmarshal(config.JSONData, &jd); err != nil {
			return nil, fmt.Errorf("could not read the datasource settings: %w", err)
		}
	}
	return &Settings{
		GrafanaURL:          strings.TrimRight(strings.TrimSpace(config.URL), "/"),
		ServiceAccountToken: strings.TrimSpace(config.DecryptedSecureJSONData["serviceAccountToken"]),
		QueryLinks:          jd.QueryLinks,
	}, nil
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
