package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonPatternAnalyzerSupportsTemporal(t *testing.T) {
	a := NewPythonPatternAnalyzer()

	temporal := []byte(`from temporalio import workflow`)
	notTemporal := []byte(`import flask`)

	if !a.Supports("file:///tmp/workflow.py", temporal) {
		t.Error("expected Supports=true for Python file with temporalio import")
	}
	if a.Supports("file:///tmp/workflow.py", notTemporal) {
		t.Error("expected Supports=false for Python file without temporalio import")
	}
	if a.Supports("file:///tmp/workflow.go", temporal) {
		t.Error("expected Supports=false for non-Go file")
	}
}

func TestPythonPatternAnalyzerDetectsBadPatterns(t *testing.T) {
	a := NewPythonPatternAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "bad_patterns.py")
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
		"temporal/activity-timeout-required",
		"temporal/unbounded-loop",
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

func TestPythonPatternAnalyzerAcceptsGoodPatterns(t *testing.T) {
	a := NewPythonPatternAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "good_patterns.py")
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
