package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRustPatternAnalyzerDetectsBadPatterns(t *testing.T) {
	a := NewRustPatternAnalyzer()

	filePath := filepath.Join(testdataDir(), "rust", "bad_patterns.rs")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	violations, err := a.Analyze("file://"+filePath, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	ruleCount := map[string]int{}
	for _, v := range violations {
		ruleCount[v.RuleID]++
	}

	if ruleCount["temporal/unbounded-loop"] == 0 {
		t.Error("expected at least one unbounded-loop violation")
	}
}

func TestRustPatternAnalyzerAcceptsGoodPatterns(t *testing.T) {
	a := NewRustPatternAnalyzer()

	filePath := filepath.Join(testdataDir(), "rust", "good_patterns.rs")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	violations, err := a.Analyze("file://"+filePath, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s: %s (line %d)", v.RuleID, v.Message, v.Range.StartLine+1)
		}
	}
}
