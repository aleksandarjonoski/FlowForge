package engine

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
)

// fakeNode is the minimal Node implementation used as a test double.
type fakeNode struct {
	typeName string
	initErr  error
	config   map[string]any
	services *EngineServices
	inited   bool
}

func (n *fakeNode) Type() string { return n.typeName }

func (n *fakeNode) Init(cfg map[string]any, svc *EngineServices) error {
	n.inited = true
	n.config = cfg
	n.services = svc
	return n.initErr
}

func newFakeFactory(typeName string) NodeFactory {
	return func() Node { return &fakeNode{typeName: typeName} }
}

func TestRegistry_RegisterAndCreate(t *testing.T) {
	r := NewRegistry()
	r.Register("test_node", newFakeFactory("test_node"))

	if !r.Has("test_node") {
		t.Error("Has(test_node) = false, want true")
	}
	if r.Has("missing") {
		t.Error("Has(missing) = true, want false")
	}

	n, err := r.Create("test_node")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got := n.Type(); got != "test_node" {
		t.Errorf("Type() = %q, want test_node", got)
	}
}

func TestRegistry_Create_UnknownType(t *testing.T) {
	r := NewRegistry()
	_, err := r.Create("nope")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnknownNodeType) {
		t.Errorf("errors.Is(err, ErrUnknownNodeType) = false; err = %v", err)
	}
}

func TestRegistry_Create_ReturnsFreshInstance(t *testing.T) {
	r := NewRegistry()
	r.Register("test_node", newFakeFactory("test_node"))

	a, _ := r.Create("test_node")
	b, _ := r.Create("test_node")
	if a == b {
		t.Error("Create returned the same instance twice; expected fresh instances")
	}

	// Mutating one instance must not affect the other.
	a.(*fakeNode).inited = true
	if b.(*fakeNode).inited {
		t.Error("mutation leaked between instances")
	}
}

func TestRegistry_Register_DuplicatePanics(t *testing.T) {
	r := NewRegistry()
	r.Register("dup", newFakeFactory("dup"))
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	r.Register("dup", newFakeFactory("dup"))
}

func TestRegistry_Register_EmptyTypePanics(t *testing.T) {
	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on empty type name")
		}
	}()
	r.Register("", newFakeFactory(""))
}

func TestRegistry_Register_NilFactoryPanics(t *testing.T) {
	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil factory")
		}
	}()
	r.Register("test", nil)
}

func TestRegistry_Types_Sorted(t *testing.T) {
	r := NewRegistry()
	r.Register("zeta", newFakeFactory("zeta"))
	r.Register("alpha", newFakeFactory("alpha"))
	r.Register("mu", newFakeFactory("mu"))

	got := r.Types()
	want := []string{"alpha", "mu", "zeta"}
	if !slices.Equal(got, want) {
		t.Errorf("Types() = %v, want %v", got, want)
	}
}

// TestRegistry_ConcurrentReads exercises Has/Create/Types under -race.
// All writes happen before goroutines spawn; only reads run in parallel.
func TestRegistry_ConcurrentReads(t *testing.T) {
	r := NewRegistry()
	for i := 0; i < 10; i++ {
		name := "t" + string(rune('a'+i))
		r.Register(name, newFakeFactory(name))
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := "t" + string(rune('a'+i%10))
			_ = r.Has(name)
			_, _ = r.Create(name)
			_ = r.Types()
		}(i)
	}
	wg.Wait()
}

// fakeAction confirms the ActionNode sub-interface is satisfiable.
type fakeAction struct{ fakeNode }

func (f *fakeAction) Execute(ctx *ExecutionContext, input Payload) (Payload, error) {
	return input, nil
}

// fakeTrigger confirms the TriggerNode sub-interface is satisfiable.
type fakeTrigger struct{ fakeNode }

func (f *fakeTrigger) Start(ctx context.Context, emit Emitter) error { return nil }
func (f *fakeTrigger) Stop() error                                   { return nil }

func TestInterfaceConformance(t *testing.T) {
	var _ Node = (*fakeNode)(nil)
	var _ ActionNode = (*fakeAction)(nil)
	var _ TriggerNode = (*fakeTrigger)(nil)
}
