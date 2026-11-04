package models

import (
	"encoding/json"
	"maps"
	"strings"
)

// Metadata decorates the data under a node of the tree: a folder or a query link
type Metadata struct {
	Description string            `json:"description,omitempty"`
	Unit        string            `json:"unit,omitempty"`
	DisplayName string            `json:"displayName,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	// Links are references such as runbooks and manufacturers' pages, shown as clickable links on the data
	Links []Reference `json:"links,omitempty"`
	Min   *float64    `json:"min,omitempty"`
	Max   *float64    `json:"max,omitempty"`
	// Thresholds are levels above which the data needs attention, lowest first
	Thresholds []Threshold `json:"thresholds,omitempty"`
	// LabelLookups add labels from queries against other datasources
	LabelLookups []LabelLookup `json:"labelLookups,omitempty"`
	// CacheTTL is how long results are reused, in seconds; 0 turns caching off for the node
	CacheTTL *int `json:"cacheTtl,omitempty"`
}

// LabelLookup is a query whose results become labels: a table's columns, or a series' labels.
// With MatchLabel, each series gets the row whose MatchColumn (default: the same name) equals
// the series' MatchLabel; without it, every series gets the first row.
type LabelLookup struct {
	UID         string          `json:"uid,omitempty"`
	Query       json.RawMessage `json:"query,omitempty"`
	MatchLabel  string          `json:"matchLabel,omitempty"`
	MatchColumn string          `json:"matchColumn,omitempty"`
}

// Datasource is the datasource the lookup queries
func (l LabelLookup) Datasource() DataSourceRef {
	return datasourceOf(l.Query, l.UID)
}

// Reference is a titled web address
type Reference struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Threshold is a level and the color data takes at or above it
type Threshold struct {
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

// IsZero reports whether there is no metadata at all
func (m Metadata) IsZero() bool {
	return m.Description == "" && m.Unit == "" && m.DisplayName == "" && len(m.Labels) == 0 &&
		len(m.Links) == 0 && m.Min == nil && m.Max == nil && len(m.Thresholds) == 0 && len(m.LabelLookups) == 0 && m.CacheTTL == nil
}

// Inherit returns m on top of parent: m's own values win, labels merge with m's winning, and
// links and label lookups accumulate, parent's first
func (m Metadata) Inherit(parent Metadata) Metadata {
	out := parent
	if m.Description != "" {
		out.Description = m.Description
	}
	if m.Unit != "" {
		out.Unit = m.Unit
	}
	if m.DisplayName != "" {
		out.DisplayName = m.DisplayName
	}
	if m.Min != nil {
		out.Min = m.Min
	}
	if m.Max != nil {
		out.Max = m.Max
	}
	if len(m.Thresholds) > 0 {
		out.Thresholds = m.Thresholds
	}
	if m.CacheTTL != nil {
		out.CacheTTL = m.CacheTTL
	}
	if len(parent.Labels)+len(m.Labels) > 0 {
		out.Labels = maps.Clone(parent.Labels)
		if out.Labels == nil {
			out.Labels = map[string]string{}
		}
		maps.Copy(out.Labels, m.Labels)
	}
	if len(m.Links) > 0 {
		out.Links = append(append([]Reference{}, parent.Links...), m.Links...)
	}
	if len(m.LabelLookups) > 0 {
		out.LabelLookups = append(append([]LabelLookup{}, parent.LabelLookups...), m.LabelLookups...)
	}
	return out
}

// MetadataFor is a link's metadata on top of the metadata of each folder above it, outermost first
func (s *Settings) MetadataFor(link QueryLink) Metadata {
	var md Metadata
	segments := strings.Split(link.Name, ".")
	for i := 1; i < len(segments); i++ {
		md = s.Folders[strings.Join(segments[:i], ".")].Inherit(md)
	}
	return link.Metadata.Inherit(md)
}
