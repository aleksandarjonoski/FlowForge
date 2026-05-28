package flow

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Parse decodes a Flow from r. It performs minimal structural checks only —
// full semantic validation (cycle detection, type registration, edge endpoint
// existence) belongs in the validator (engine v1 §4.2).
func Parse(r io.Reader) (*Flow, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var f Flow
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("flow: decode: %w", err)
	}
	if err := basicChecks(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

// ParseFile is a convenience wrapper around Parse for the common case of
// loading a flow from disk.
func ParseFile(path string) (*Flow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("flow: open %s: %w", path, err)
	}
	defer file.Close()
	return Parse(file)
}

// basicChecks enforces only the rules that are cheap and unambiguous at parse
// time. Anything that requires the node registry or graph analysis lives in
// the validator.
func basicChecks(f *Flow) error {
	if f.ID == "" {
		return fmt.Errorf("flow: id is required")
	}
	if f.Version == "" {
		return fmt.Errorf("flow: version is required")
	}
	if f.Version != SchemaVersion {
		return fmt.Errorf("flow: unsupported version %q (this build understands %q)", f.Version, SchemaVersion)
	}
	if len(f.Nodes) == 0 {
		return fmt.Errorf("flow: must contain at least one node")
	}

	seen := make(map[string]struct{}, len(f.Nodes))
	for i, n := range f.Nodes {
		if n.ID == "" {
			return fmt.Errorf("flow: nodes[%d]: id is required", i)
		}
		if n.Type == "" {
			return fmt.Errorf("flow: nodes[%d] (%s): type is required", i, n.ID)
		}
		if _, dup := seen[n.ID]; dup {
			return fmt.Errorf("flow: duplicate node id %q", n.ID)
		}
		seen[n.ID] = struct{}{}
	}

	edgeIDs := make(map[string]struct{}, len(f.Edges))
	for i, e := range f.Edges {
		if e.Source == "" || e.Target == "" {
			return fmt.Errorf("flow: edges[%d]: source and target are required", i)
		}
		if e.ID != "" {
			if _, dup := edgeIDs[e.ID]; dup {
				return fmt.Errorf("flow: duplicate edge id %q", e.ID)
			}
			edgeIDs[e.ID] = struct{}{}
		}
	}
	return nil
}
