package models

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// QueryLinkConfig ...
type QueryLinkConfig struct {
	Name  string                 `json:"name"`
	Query map[string]interface{} `json:"query,omitempty"`
	UID   string                 `json:"uid"`
}

type TypeInfo struct {
	Frame    string `json:"frame,omitempty"`
	Nullable bool   `json:"nullable,omitempty"`
}

type Field struct {
	Name     string            `json:"name,omitempty"`
	Labels   map[string]string `json:"config,omitempty"`
	State    *data.FieldConfig `json:"state,omitempty"`
	Type     string            `json:"type,omitempty"`
	TypeInfo TypeInfo          `json:"typeinfo,omitempty"`
	Values   []interface{}     `json:"values,omitempty"`
}

type Frame struct {
	// Name is used in some Grafana visualizations.
	Name string `json:"name"`

	// Fields are the columns of a frame.
	// All Fields must be of the same the length when marshalling the Frame for transmission.
	Fields []*Field

	// RefID is a property that can be set to match a Frame to its orginating query.
	RefID string

	// Meta is metadata about the Frame, and includes space for custom metadata.
	Meta *data.FrameMeta
}

type DataQueryResponse struct {
	Data []*Frame `json:"data"`
}

// ObjectDataQuery ...
type ObjectDataQuery struct {
	Name     string            `json:"name"`
	Config   QueryLinkConfig   `json:"config"`
	Response DataQueryResponse `json:"response"`
}

type ProxiedResponse struct {
	// Responses is a map of RefIDs (Unique Query ID) to *DataResponse.
	Responses ProxiedResponses `json:"results"`
}

type ProxiedResponses map[string]ProxiedDataResponse

type ProxiedDataResponse struct {
	Frames []*data.Frame `json:"frames"`
}

type ProxiedTimeSeriesDataResponse struct {
	RefID      string   `json:"refId"`
	Series     [][]byte `json:"series,omitempty"`
	Tabes      [][]byte `json:"tables,omitempty"`
	DataFrames [][]byte `json:"dataframes,omitempty"`
}
