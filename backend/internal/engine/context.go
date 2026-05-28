package engine

import "context"

// ExecutionContext is the per-execution state passed to every ActionNode.
// See docs/engine-v1.md §3.3 for the full contract. This minimal version is
// in place so the ActionNode interface can be locked from slice 3; the
// executor (slice 5) populates these fields and may add helper methods.
type ExecutionContext struct {
	FlowID      string
	ExecutionID string
	NodeResults map[string]Payload
	Trace       TraceSink
	GoCtx       context.Context
}
