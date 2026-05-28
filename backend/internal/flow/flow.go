// Package flow defines the FlowForge workflow document model and its JSON
// serialization. See docs/engine-v1.md §3.4 for the contract.
package flow

// SchemaVersion is the version string this package understands.
const SchemaVersion = "1.0"

// Flow is the top-level workflow document. It mirrors the v1 JSON schema 1:1.
type Flow struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Nodes    []Node   `json:"nodes"`
	Edges    []Edge   `json:"edges"`
	Metadata Metadata `json:"metadata"`
}

// Node is a single block in the graph. Config is intentionally untyped at this
// layer — each node implementation defines its own config contract.
type Node struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	Position Position       `json:"position"`
	Config   map[string]any `json:"config"`
}

// Edge connects two nodes. SourceHandle and TargetHandle are reserved for v2;
// the v1 engine ignores them.
type Edge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	SourceHandle string `json:"sourceHandle,omitempty"`
	Target       string `json:"target"`
	TargetHandle string `json:"targetHandle,omitempty"`
}

// Position is the node's coordinate in the visual editor. The runtime does
// not use these values, but they are preserved through load/save round-trips.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Metadata holds bookkeeping fields that are not part of execution semantics.
type Metadata struct {
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Author    string `json:"author,omitempty"`
}
