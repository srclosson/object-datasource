package models

import "encoding/json"

// ProxiedDataRequest uses the grafana http datasource api to proxy datasource queries to backends
type ProxiedDataRequest struct {
	From    string            `json:"from"`
	To      string            `json:"to"`
	Queries []json.RawMessage `json:"queries"`
}

// NewQueryDataResponse returns a QueryDataResponse with the Responses property initialized.
// func NewQueryDataResponse() *QueryDataResponse {
// 	return &QueryDataResponse{
// 		Responses: make(Responses),
// 	}
// }

// Responses is a map of RefIDs (Unique Query ID) to DataResponses.
// The QueryData method the QueryDataHandler method will set the RefId
// property on the DataRespones' frames based on these RefIDs.
// type Responses map[string]DataResponse `json:`

// DataResponse contains the results from a DataQuery.
// A map of RefIDs (unique query identifers) to this type makes up the Responses property of a QueryDataResponse.
// The Error property is used to allow for partial success responses from the containing QueryDataResponse.
// type DataResponse struct {
// 	// The data returned from the Query. Each Frame repeats the RefID.
// 	Frames data.Frames

// 	// Error is a property to be set if the the corresponding DataQuery has an error.
// 	Error error
// }
