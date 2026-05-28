package engine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
)

// Payload is the data that flows between nodes. It is map[string]any so the
// schema-less node config philosophy is consistent at runtime, and so JSON
// serialization (for tracing and persistence) is a one-liner.
type Payload = map[string]any

// Emitter is what a TriggerNode calls to fire an event. Each call spawns a
// new execution of the trigger's downstream subgraph.
type Emitter func(Payload) error

// Node is the base interface every node type implements. See
// docs/engine-v1.md §3.2 + §12.
//
// A concrete type must additionally satisfy either ActionNode or
// TriggerNode. The compiler enforces this when nodes are instantiated.
type Node interface {
	Type() string
	Init(config map[string]any, services *EngineServices) error
}

// ActionNode is invoked synchronously during an execution. Input is the
// upstream node's output (or the trigger's emitted payload, for the first
// action after a trigger).
type ActionNode interface {
	Node
	Execute(ctx *ExecutionContext, input Payload) (Payload, error)
}

// TriggerNode runs for the lifetime of the flow. It calls emit whenever an
// event occurs; the engine spawns one execution per call.
type TriggerNode interface {
	Node
	Start(ctx context.Context, emit Emitter) error
	Stop() error
}

// NodeFactory produces a fresh node instance. The registry calls the factory
// once per occurrence in the flow, so two nodes of the same type get
// independent state.
type NodeFactory func() Node

// ErrUnknownNodeType is the sentinel error returned by Create for a type
// name that has not been registered. Callers may use errors.Is to detect it.
var ErrUnknownNodeType = errors.New("unknown node type")

// Registry maps node type names to factories. The package-level
// DefaultRegistry is what built-in nodes self-register against in their
// init() functions; tests and embedded uses can create their own with
// NewRegistry for isolation.
//
// Safe for concurrent use.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]NodeFactory
}

// NewRegistry returns a new, empty Registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]NodeFactory)}
}

// Register adds a node type to the registry.
//
// Panics if typeName is empty, the factory is nil, or the type is already
// registered. Registration is a startup-time concern — a duplicate usually
// means two init() functions for the same type, which is a bug that should
// fail loudly rather than silently win or lose.
func (r *Registry) Register(typeName string, f NodeFactory) {
	if typeName == "" {
		panic("engine: Register called with empty type name")
	}
	if f == nil {
		panic("engine: Register called with nil factory for " + typeName)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[typeName]; exists {
		panic("engine: node type already registered: " + typeName)
	}
	r.factories[typeName] = f
}

// Create instantiates a new node of the given type. Returns
// ErrUnknownNodeType (wrapped) if the type is not registered, so callers can
// surface a precise error with the failing flow + node ID context.
func (r *Registry) Create(typeName string) (Node, error) {
	r.mu.RLock()
	f, ok := r.factories[typeName]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownNodeType, typeName)
	}
	return f(), nil
}

// Has reports whether a node type is registered.
func (r *Registry) Has(typeName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.factories[typeName]
	return ok
}

// Types returns all registered type names in lexicographic order. Useful for
// diagnostics and for populating the node palette in the UI.
func (r *Registry) Types() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.factories))
	for k := range r.factories {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// DefaultRegistry is the process-wide registry that built-in nodes
// self-register against via init(). Application code may use it directly,
// but tests should prefer NewRegistry() for isolation.
var DefaultRegistry = NewRegistry()
