package models

import (
	"encoding/json"
	"fmt"
)

// ObjectQuery is a panel's or alert rule's query: a reference to a query link, which is
// resolved from the datasource settings each time it runs
type ObjectQuery struct {
	LinkID string `json:"linkId,omitempty"`
	// Name is the link's path in the tree, kept for display and as a fallback for links
	// saved before they had ids
	Name string `json:"name,omitempty"`
	// Variables are the values of the dashboard variables the link's query uses
	Variables map[string]VarValue `json:"variables,omitempty"`
}

// VarValue is a template variable's value: one value, or several for a multi-value variable
type VarValue []string

// UnmarshalJSON accepts a string or an array of strings
func (v *VarValue) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*v = VarValue{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return fmt.Errorf("a variable's value must be a string or an array of strings")
	}
	*v = many
	return nil
}
