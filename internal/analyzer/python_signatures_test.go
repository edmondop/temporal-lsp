package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonSignatureAnalyzerSupportsTemporal(t *testing.T) {
	a := NewPythonSignatureAnalyzer()

	temporal := []byte(`from temporalio import workflow`)
	notTemporal := []byte(`import flask`)

	if !a.Supports("file:///tmp/workflow.py", temporal) {
		t.Error("expected Supports=true for Python file with temporalio import")
	}
	if a.Supports("file:///tmp/workflow.py", notTemporal) {
		t.Error("expected Supports=false for Python file without temporalio import")
	}
	if a.Supports("file:///tmp/workflow.go", temporal) {
		t.Error("expected Supports=false for non-Python file")
	}
}

func TestPythonSignatureAnalyzerDetectsBadSignatures(t *testing.T) {
	a := NewPythonSignatureAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "bad_signatures.py")
	uri := "file://" + filePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	violations, err := a.Analyze(uri, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(violations) < 2 {
		t.Errorf("expected at least 2 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s: %s (line %d)", v.RuleID, v.Message, v.Range.StartLine+1)
		}
	}

	ruleCount := map[ID]int{}
	for _, v := range violations {
		ruleCount[v.RuleID]++
	}

	// Both the workflow.run method and the activity should trigger single-payload
	if ruleCount["temporal/single-payload"] < 2 {
		t.Errorf("expected at least 2 single-payload violations, got %d", ruleCount["temporal/single-payload"])
	}

	// Both should also trigger primitive-params (multiple str/int/bool params)
	if ruleCount["temporal/primitive-params"] < 2 {
		t.Errorf("expected at least 2 primitive-params violations, got %d", ruleCount["temporal/primitive-params"])
	}
}

func TestPythonSignatureAnalyzerAcceptsGoodSignatures(t *testing.T) {
	a := NewPythonSignatureAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "good_signatures.py")
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
