package analyzer

import "sync"

// Registry holds the registered WasteAnalyzers.
type Registry interface {
	Register(analyzer WasteAnalyzer)
	GetAnalyzers() []WasteAnalyzer
}

type registry struct {
	analyzers []WasteAnalyzer
	mu        sync.RWMutex
}

// NewRegistry creates a new WasteAnalyzer registry.
func NewRegistry() Registry {
	return &registry{
		analyzers: make([]WasteAnalyzer, 0),
	}
}

func (r *registry) Register(analyzer WasteAnalyzer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.analyzers = append(r.analyzers, analyzer)
}

func (r *registry) GetAnalyzers() []WasteAnalyzer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent race conditions during iteration
	result := make([]WasteAnalyzer, len(r.analyzers))
	copy(result, r.analyzers)

	return result
}
