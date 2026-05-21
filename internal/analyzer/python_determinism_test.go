package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonDeterminismAnalyzerSupportsTemporal(t *testing.T) {
	a := NewPythonDeterminismAnalyzer()

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

func TestPythonDeterminismAnalyzerDetectsBannedCalls(t *testing.T) {
	a := NewPythonDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "bad_workflow.py")
	uri := "file://" + filePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	violations, err := a.Analyze(uri, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(violations) < 7 {
		t.Errorf("expected at least 7 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s: %s (line %d)", v.RuleID, v.Message, v.Range.StartLine+1)
		}
	}

	ruleCount := map[ID]int{}
	for _, v := range violations {
		ruleCount[v.RuleID]++
	}

	expected := map[ID]bool{
		"temporal/no-time-now":  true,
		"temporal/no-sleep":    true,
		"temporal/no-random":   true,
		"temporal/no-io":       true,
		"temporal/no-goroutine": true,
		"temporal/no-mutex":    true,
		"temporal/no-channel":  true,
	}
	for rule := range expected {
		if ruleCount[rule] == 0 {
			t.Errorf("expected at least one %s violation", rule)
		}
	}
}

func TestPythonDeterminismAnalyzerHandlesDirectImport(t *testing.T) {
	a := NewPythonDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "bad_workflow_import_variants.py")
	uri := "file://" + filePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	if !a.Supports(uri, content) {
		t.Fatal("expected Supports=true for file with temporalio.workflow import")
	}

	violations, err := a.Analyze(uri, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(violations) < 2 {
		t.Errorf("expected at least 2 violations (no-time-now, no-sleep), got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s: %s (line %d)", v.RuleID, v.Message, v.Range.StartLine+1)
		}
	}

	ruleCount := map[ID]int{}
	for _, v := range violations {
		ruleCount[v.RuleID]++
	}
	if ruleCount["temporal/no-time-now"] == 0 {
		t.Error("expected no-time-now violation")
	}
	if ruleCount["temporal/no-sleep"] == 0 {
		t.Error("expected no-sleep violation")
	}
}

func TestPythonDeterminismAnalyzerDetectsEnvAccess(t *testing.T) {
	a := NewPythonDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "bad_env_workflow.py")
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
	if ruleCount["temporal/no-env-access"] < 2 {
		t.Errorf("expected at least 2 no-env-access violations, got %d", ruleCount["temporal/no-env-access"])
		for _, v := range violations {
			t.Logf("  %s: %s (line %d)", v.RuleID, v.Message, v.Range.StartLine+1)
		}
	}
}

func TestPythonDeterminismAnalyzerDetectsStandardLogging(t *testing.T) {
	a := NewPythonDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "bad_logging_workflow.py")
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
	if ruleCount["temporal/no-standard-logging"] < 3 {
		t.Errorf("expected at least 3 no-standard-logging violations, got %d", ruleCount["temporal/no-standard-logging"])
		for _, v := range violations {
			t.Logf("  %s: %s (line %d)", v.RuleID, v.Message, v.Range.StartLine+1)
		}
	}
}

func TestPythonDeterminismAnalyzerAcceptsGoodWorkflow(t *testing.T) {
	a := NewPythonDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "python", "good_workflow.py")
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
