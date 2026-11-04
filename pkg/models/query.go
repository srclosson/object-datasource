package models

// ObjectQuery is a panel's or alert rule's query: a reference to a query link, which is
// resolved from the datasource settings each time it runs
type ObjectQuery struct {
	LinkID string `json:"linkId,omitempty"`
	// Name is the link's path in the tree, kept for display and as a fallback for links
	// saved before they had ids
	Name string `json:"name,omitempty"`
}
