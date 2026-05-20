package analyzer

import "testing"

func TestAllAnalyzersReturnsCorrectCount(t *testing.T) {
	analyzers := AllAnalyzers(Go{}, Python{}, Java{}, Rust{})

	// 3 analyzers per language × 4 languages = 12
	if len(analyzers) != 12 {
		t.Errorf("expected 12 analyzers, got %d", len(analyzers))
	}
}
