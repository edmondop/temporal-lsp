package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJavaPatternAnalyzerDetectsBadPatterns(t *testing.T) {
	a := NewJavaPatternAnalyzer()

	filePath := filepath.Join(testdataDir(), "java", "BadPatterns.java")
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

func TestJavaPatternAnalyzerAcceptsGoodPatterns(t *testing.T) {
	a := NewJavaPatternAnalyzer()

	filePath := filepath.Join(testdataDir(), "java", "GoodPatterns.java")
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
