package flow

import (
	"strings"
	"testing"
)

func TestParseFile_HelloExample(t *testing.T) {
	f, err := ParseFile("../../examples/hello.flow.json")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if f.ID != "flow-hello" {
		t.Errorf("ID = %q, want %q", f.ID, "flow-hello")
	}
	if f.Version != SchemaVersion {
		t.Errorf("Version = %q, want %q", f.Version, SchemaVersion)
	}
	if got := len(f.Nodes); got != 3 {
		t.Fatalf("len(Nodes) = %d, want 3", got)
	}
	if got := len(f.Edges); got != 2 {
		t.Fatalf("len(Edges) = %d, want 2", got)
	}
	if f.Nodes[0].Type != "http_trigger" {
		t.Errorf("Nodes[0].Type = %q", f.Nodes[0].Type)
	}
	if f.Nodes[0].Config["path"] != "/webhook" {
		t.Errorf("http_trigger path = %v, want /webhook", f.Nodes[0].Config["path"])
	}
	if f.Edges[0].Source != "node-1" || f.Edges[0].Target != "node-2" {
		t.Errorf("Edges[0] = %+v", f.Edges[0])
	}
}

func TestParse_Errors(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantSub string
	}{
		{
			name:    "invalid JSON",
			input:   `{not json`,
			wantSub: "decode",
		},
		{
			name:    "missing id",
			input:   `{"version":"1.0","nodes":[{"id":"n","type":"log","position":{"x":0,"y":0}}]}`,
			wantSub: "id is required",
		},
		{
			name:    "missing version",
			input:   `{"id":"f","nodes":[{"id":"n","type":"log","position":{"x":0,"y":0}}]}`,
			wantSub: "version is required",
		},
		{
			name:    "unsupported version",
			input:   `{"id":"f","version":"9.9","nodes":[{"id":"n","type":"log","position":{"x":0,"y":0}}]}`,
			wantSub: "unsupported version",
		},
		{
			name:    "no nodes",
			input:   `{"id":"f","version":"1.0","nodes":[]}`,
			wantSub: "at least one node",
		},
		{
			name:    "duplicate node id",
			input:   `{"id":"f","version":"1.0","nodes":[{"id":"n","type":"log","position":{"x":0,"y":0}},{"id":"n","type":"log","position":{"x":0,"y":0}}]}`,
			wantSub: "duplicate node id",
		},
		{
			name:    "node missing type",
			input:   `{"id":"f","version":"1.0","nodes":[{"id":"n","position":{"x":0,"y":0}}]}`,
			wantSub: "type is required",
		},
		{
			name:    "edge missing endpoints",
			input:   `{"id":"f","version":"1.0","nodes":[{"id":"n","type":"log","position":{"x":0,"y":0}}],"edges":[{"id":"e","source":""}]}`,
			wantSub: "source and target are required",
		},
		{
			name:    "duplicate edge id",
			input:   `{"id":"f","version":"1.0","nodes":[{"id":"a","type":"log","position":{"x":0,"y":0}},{"id":"b","type":"log","position":{"x":0,"y":0}}],"edges":[{"id":"e","source":"a","target":"b"},{"id":"e","source":"b","target":"a"}]}`,
			wantSub: "duplicate edge id",
		},
		{
			name:    "unknown field",
			input:   `{"id":"f","version":"1.0","extra":1,"nodes":[{"id":"n","type":"log","position":{"x":0,"y":0}}]}`,
			wantSub: "unknown field",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tc.input))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("error = %q, want substring %q", err.Error(), tc.wantSub)
			}
		})
	}
}

func TestParse_PreservesUntypedConfig(t *testing.T) {
	in := `{
	  "id":"f","version":"1.0",
	  "nodes":[{"id":"n","type":"transform","position":{"x":0,"y":0},
	    "config":{"expression":"input.x + 1","nested":{"a":[1,2,3]}}}]
	}`
	f, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	cfg := f.Nodes[0].Config
	if cfg["expression"] != "input.x + 1" {
		t.Errorf("expression = %v", cfg["expression"])
	}
	nested, ok := cfg["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested not a map: %T", cfg["nested"])
	}
	arr, ok := nested["a"].([]any)
	if !ok || len(arr) != 3 {
		t.Errorf("nested.a = %v", nested["a"])
	}
}
