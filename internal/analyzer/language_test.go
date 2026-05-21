package analyzer

import "testing"

func TestAllAnalyzersReturnsCorrectCount(t *testing.T) {
	analyzers := AllAnalyzers()

	// 3 analyzers per language × 6 languages = 18
	if len(analyzers) != 18 {
		t.Errorf("expected 18 analyzers, got %d", len(analyzers))
	}
}
