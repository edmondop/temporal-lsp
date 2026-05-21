package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoPatternAnalyzerSupportsTemporal(t *testing.T) {
	a := NewGoPatternAnalyzer()

	temporal := []byte(`package foo
import "go.temporal.io/sdk/workflow"
`)
	notTemporal := []byte(`package foo
import "net/http"
`)

	if !a.Supports("file:///tmp/workflow.go", temporal) {
		t.Error("expected Supports=true for Go file with workflow import")
	}
	if a.Supports("file:///tmp/workflow.go", notTemporal) {
		t.Error("expected Supports=false for Go file without temporal import")
	}
	if a.Supports("file:///tmp/workflow.py", temporal) {
		t.Error("expected Supports=false for non-Go file")
	}
}

func TestGoPatternAnalyzerDetectsBadPatterns(t *testing.T) {
	a := NewGoPatternAnalyzer()

	filePath := filepath.Join(testdataDir(), "go", "patterns", "bad.go")
	uri := "file://" + filePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	violations, err := a.Analyze(uri, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	ruleCount := map[ID]int{}
	for _, v := range violations {
		ruleCount[v.RuleID]++
	}

	expected := []ID{
		"temporal/no-context-propagation",
		"temporal/activity-timeout-required",
		"temporal/unbounded-loop",
		"temporal/no-naked-error",
	}
	for _, rule := range expected {
		if ruleCount[rule] == 0 {
			t.Errorf("expected at least one %s violation, got none", rule)
			t.Logf("all violations:")
			for _, v := range violations {
				t.Logf("  %s: %s (line %d)", v.RuleID, v.Message, v.Range.StartLine+1)
			}
		}
	}
}

func TestGoPatternAnalyzerAcceptsGoodPatterns(t *testing.T) {
	a := NewGoPatternAnalyzer()

	filePath := filepath.Join(testdataDir(), "go", "patterns", "good.go")
	uri := "file://" + filePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	violations, err := a.Analyze(uri, content)
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
