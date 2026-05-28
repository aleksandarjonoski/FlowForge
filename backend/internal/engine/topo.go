// Package engine is the FlowForge execution runtime. It compiles a parsed
// Flow into an executable form, validates the graph, and runs it. See
// docs/engine-v1.md for the full contract.
//
// This file contains pure graph algorithms operating on adjacency maps —
// cycle detection, topological sort, reachability. They are independent of
// the Flow / Node types so they can be tested in isolation and reused by
// the validator and the per-trigger execution planner.
package engine

import (
	"fmt"
	"slices"
)

// adjacency is the input shape every algorithm in this file accepts.
// adj[u] is the list of nodes that u has outgoing edges to. A node with
// no outgoing edges may either be absent or map to a nil/empty slice.
//
// Nodes that appear only as edge targets (not as keys) are still considered
// part of the graph.
type adjacency = map[string][]string

// detectCycle walks the graph looking for any cycle. If one is found, it
// returns the cycle as a list of node IDs in traversal order: the first and
// last entries are the same node so a cycle A -> B -> C -> A is returned as
// ["A", "B", "C", "A"]. A self-loop A -> A is returned as ["A", "A"].
//
// Iteration is deterministic: nodes and neighbors are processed in
// lexicographic order, so the same input always yields the same reported
// cycle. Useful for stable error messages and reproducible tests.
func detectCycle(adj adjacency) ([]string, bool) {
	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[string]int)
	parent := make(map[string]string)
	var cycleHead, cycleTail string

	var dfs func(u string) bool
	dfs = func(u string) bool {
		color[u] = gray
		neighbors := append([]string(nil), adj[u]...)
		slices.Sort(neighbors)
		for _, v := range neighbors {
			switch color[v] {
			case white:
				parent[v] = u
				if dfs(v) {
					return true
				}
			case gray:
				// Back-edge u -> v closes a cycle. v is the head of the
				// cycle; u is the tail (the node that closes it).
				cycleHead, cycleTail = v, u
				return true
			}
		}
		color[u] = black
		return false
	}

	for _, n := range allNodes(adj) {
		if color[n] != white {
			continue
		}
		if dfs(n) {
			return reconstructCycle(parent, cycleHead, cycleTail), true
		}
	}
	return nil, false
}

// reconstructCycle walks the parent map from tail back to head, then reverses
// to produce the forward path, then appends head to close the cycle.
func reconstructCycle(parent map[string]string, head, tail string) []string {
	path := []string{tail}
	for v := tail; v != head; {
		v = parent[v]
		path = append(path, v)
	}
	slices.Reverse(path)
	return append(path, head)
}

// topoSort returns the nodes of adj in topological order using Kahn's
// algorithm. If only is non-nil, only nodes in the set are considered, and
// edges to or from nodes outside the set are ignored — useful for executing
// just the subgraph reachable from a particular trigger.
//
// Returns an error if the (restricted) graph contains a cycle. Ties are
// broken by node ID in lexicographic order so the result is deterministic.
func topoSort(adj adjacency, only map[string]bool) ([]string, error) {
	inSet := func(n string) bool { return only == nil || only[n] }

	nodes := allNodes(adj)
	nodes = slices.DeleteFunc(nodes, func(n string) bool { return !inSet(n) })

	inDeg := make(map[string]int, len(nodes))
	for _, n := range nodes {
		inDeg[n] = 0
	}
	for _, u := range nodes {
		for _, v := range adj[u] {
			if inSet(v) {
				inDeg[v]++
			}
		}
	}

	var ready []string
	for n, d := range inDeg {
		if d == 0 {
			ready = append(ready, n)
		}
	}
	slices.Sort(ready)

	result := make([]string, 0, len(nodes))
	for len(ready) > 0 {
		u := ready[0]
		ready = ready[1:]
		result = append(result, u)

		neighbors := append([]string(nil), adj[u]...)
		slices.Sort(neighbors)
		for _, v := range neighbors {
			if !inSet(v) {
				continue
			}
			inDeg[v]--
			if inDeg[v] == 0 {
				idx, _ := slices.BinarySearch(ready, v)
				ready = slices.Insert(ready, idx, v)
			}
		}
	}

	if len(result) != len(nodes) {
		return nil, fmt.Errorf("topo sort: cycle detected (processed %d of %d nodes)", len(result), len(nodes))
	}
	return result, nil
}

// reachable returns the set of nodes reachable from `from` by following
// outgoing edges. The result includes `from` itself. Safe on cyclic graphs.
func reachable(adj adjacency, from string) map[string]bool {
	visited := make(map[string]bool)
	stack := []string{from}
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[u] {
			continue
		}
		visited[u] = true
		for _, v := range adj[u] {
			if !visited[v] {
				stack = append(stack, v)
			}
		}
	}
	return visited
}

// allNodes returns the union of keys and edge targets in adj, sorted
// lexicographically. Nodes that appear only as edge targets are included.
func allNodes(adj adjacency) []string {
	set := make(map[string]struct{}, len(adj))
	for u, neighbors := range adj {
		set[u] = struct{}{}
		for _, v := range neighbors {
			set[v] = struct{}{}
		}
	}
	nodes := make([]string, 0, len(set))
	for n := range set {
		nodes = append(nodes, n)
	}
	slices.Sort(nodes)
	return nodes
}
