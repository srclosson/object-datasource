package models

// QueryLinkConfig ...
type QueryLinkConfig struct {
	Name  string                 `json:"name"`
	Query map[string]interface{} `json:"query,omitempty"`
	UID   string                 `json:"uid"`
}

// ObjectDataQuery ...
type ObjectDataQuery struct {
	Name   string          `json:"name"`
	Config QueryLinkConfig `json:"config"`
}

type ProxiedResponse struct {
	// Responses is a map of RefIDs (Unique Query ID) to *DataResponse.
	Responses ProxiedResponses `json:"results"`
}

type ProxiedResponses map[string]ProxiedDataResponse

type ProxiedDataResponse struct {
	RefID      string   `json:"refId"`
	Series     [][]byte `json:"series,omitempty"`
	Tabes      [][]byte `json:"tables,omitempty"`
	DataFrames [][]byte `json:"dataframes,omitempty"`
}
