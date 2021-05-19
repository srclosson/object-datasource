package models

import "encoding/json"

type QueryLinkConfig struct {
	Query json.RawMessage `json:"query,omitempty"`
}

type ObjectDataQuery struct {
	Name   string          `json:"name"`
	Config QueryLinkConfig `json:"config"`
}
