package engine

// EngineServices holds shared infrastructure that some nodes need at runtime:
// the HTTP server (for http_trigger), the cron scheduler (for cron_trigger),
// etc. See docs/engine-v1.md §12.
//
// Concrete fields are added by their respective slices (HTTP: slice 7, cron:
// slice 9). This stub exists so the Node.Init signature can be locked from
// slice 3 onward without future breaking changes.
type EngineServices struct {
	// Reserved for future fields.
}
