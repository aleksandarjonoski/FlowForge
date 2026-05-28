package engine

import (
	"slices"
	"strings"
	"testing"
)

func TestDetectCycle(t *testing.T) {
	cases := []struct {
		name      string
		adj       adjacency
		wantCycle []string // nil = no cycle expected
	}{
		{
			name: "empty graph",
			adj:  adjacency{},
		},
		{
			name: "single isolated node",
			adj:  adjacency{"A": nil},
		},
		{
			name: "linear chain",
			adj:  adjacency{"A": {"B"}, "B": {"C"}, "C": nil},
		},
		{
			name: "diamond DAG",
			adj:  adjacency{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": nil},
		},
		{
			name: "disconnected DAG components",
			adj:  adjacency{"A": {"B"}, "B": nil, "X": {"Y"}, "Y": nil},
		},
		{
			name:      "3-node cycle",
			adj:       adjacency{"A": {"B"}, "B": {"C"}, "C": {"A"}},
			wantCycle: []string{"A", "B", "C", "A"},
		},
		{
			name:      "self-loop",
			adj:       adjacency{"A": {"A"}},
			wantCycle: []string{"A", "A"},
		},
		{
			name:      "2-node cycle",
			adj:       adjacency{"A": {"B"}, "B": {"A"}},
			wantCycle: []string{"A", "B", "A"},
		},
		{
			// A -> B -> C -> B. DFS starts at A, walks to B then C, then
			// finds back-edge C -> B. The reported cycle is the inner loop
			// only, not the lead-in from A.
			name:      "cycle reachable from outside",
			adj:       adjacency{"A": {"B"}, "B": {"C"}, "C": {"B"}},
			wantCycle: []string{"B", "C", "B"},
		},
		{
			// Two disconnected components; the second has the cycle.
			// DFS processes A's component first (clean), then X's.
			name:      "cycle in second component",
			adj:       adjacency{"A": {"B"}, "B": nil, "X": {"Y"}, "Y": {"X"}},
			wantCycle: []string{"X", "Y", "X"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cycle, found := detectCycle(tc.adj)
			wantFound := tc.wantCycle != nil
			if found != wantFound {
				t.Fatalf("found = %v, want %v (returned cycle=%v)", found, wantFound, cycle)
			}
			if found && !slices.Equal(cycle, tc.wantCycle) {
				t.Errorf("cycle = %v, want %v", cycle, tc.wantCycle)
			}
		})
	}
}

func TestTopoSort(t *testing.T) {
	cases := []struct {
		name    string
		adj     adjacency
		only    map[string]bool
		want    []string
		wantErr string
	}{
		{
			name: "empty graph",
			adj:  adjacency{},
			want: []string{},
		},
		{
			name: "linear chain",
			adj:  adjacency{"A": {"B"}, "B": {"C"}, "C": nil},
			want: []string{"A", "B", "C"},
		},
		{
			// A has two children; ties broken alphabetically (B before C),
			// then D last because it depends on both.
			name: "diamond, ties broken alphabetically",
			adj:  adjacency{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": nil},
			want: []string{"A", "B", "C", "D"},
		},
		{
			// Two roots A and X; alphabetical order means A's component
			// finishes before X's begins (because B becomes ready before X
			// is processed only if A is processed first).
			name: "disconnected components",
			adj:  adjacency{"A": {"B"}, "B": nil, "X": {"Y"}, "Y": nil},
			want: []string{"A", "B", "X", "Y"},
		},
		{
			name:    "cycle returns error",
			adj:     adjacency{"A": {"B"}, "B": {"A"}},
			wantErr: "cycle detected",
		},
		{
			// A is excluded; the subgraph is B -> C -> D.
			name: "only restricts to subset",
			adj:  adjacency{"A": {"B"}, "B": {"C"}, "C": {"D"}, "D": nil},
			only: map[string]bool{"B": true, "C": true, "D": true},
			want: []string{"B", "C", "D"},
		},
		{
			// C is excluded; A -> C and C -> D are ignored, so the
			// remaining subgraph is just A -> B -> D.
			name: "only excludes some edges",
			adj:  adjacency{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": nil},
			only: map[string]bool{"A": true, "B": true, "D": true},
			want: []string{"A", "B", "D"},
		},
		{
			// Singleton in `only`: node has no inbound or outbound edges
			// within the restricted subgraph.
			name: "only single node",
			adj:  adjacency{"A": {"B"}, "B": nil},
			only: map[string]bool{"A": true},
			want: []string{"A"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := topoSort(tc.adj, tc.only)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (result=%v)", tc.wantErr, got)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error = %q, want substring %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("topoSort = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestReachable(t *testing.T) {
	cases := []struct {
		name string
		adj  adjacency
		from string
		want []string
	}{
		{
			name: "isolated node",
			adj:  adjacency{"A": nil},
			from: "A",
			want: []string{"A"},
		},
		{
			name: "linear chain from start",
			adj:  adjacency{"A": {"B"}, "B": {"C"}, "C": nil},
			from: "A",
			want: []string{"A", "B", "C"},
		},
		{
			name: "linear chain from middle",
			adj:  adjacency{"A": {"B"}, "B": {"C"}, "C": nil},
			from: "B",
			want: []string{"B", "C"},
		},
		{
			name: "diamond",
			adj:  adjacency{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": nil},
			from: "A",
			want: []string{"A", "B", "C", "D"},
		},
		{
			name: "disconnected component is unreachable",
			adj:  adjacency{"A": {"B"}, "B": nil, "X": {"Y"}, "Y": nil},
			from: "A",
			want: []string{"A", "B"},
		},
		{
			// Demonstrates that reachable does not loop forever on cycles.
			name: "cycle is traversed safely",
			adj:  adjacency{"A": {"B"}, "B": {"A"}},
			from: "A",
			want: []string{"A", "B"},
		},
		{
			// Node not in the adjacency map at all; still trivially
			// reachable from itself.
			name: "unknown start node",
			adj:  adjacency{"A": {"B"}, "B": nil},
			from: "Z",
			want: []string{"Z"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := reachable(tc.adj, tc.from)
			keys := make([]string, 0, len(got))
			for k := range got {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			if !slices.Equal(keys, tc.want) {
				t.Errorf("reachable = %v, want %v", keys, tc.want)
			}
		})
	}
}

func TestAllNodes(t *testing.T) {
	cases := []struct {
		name string
		adj  adjacency
		want []string
	}{
		{
			name: "empty",
			adj:  adjacency{},
			want: []string{},
		},
		{
			name: "key only",
			adj:  adjacency{"A": nil},
			want: []string{"A"},
		},
		{
			// B is referenced as a target but never appears as a key.
			// allNodes must still include it.
			name: "target without key",
			adj:  adjacency{"A": {"B"}},
			want: []string{"A", "B"},
		},
		{
			name: "deduped union",
			adj:  adjacency{"A": {"B", "C"}, "B": {"C"}},
			want: []string{"A", "B", "C"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := allNodes(tc.adj)
			if !slices.Equal(got, tc.want) {
				t.Errorf("allNodes = %v, want %v", got, tc.want)
			}
		})
	}
}
