package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJavaDeterminismAnalyzerSupportsTemporalFiles(t *testing.T) {
	a := NewJavaDeterminismAnalyzer()

	temporal := []byte(`import io.temporal.workflow.WorkflowMethod;`)
	notTemporal := []byte(`import java.util.List;`)

	if !a.Supports("file:///tmp/Workflow.java", temporal) {
		t.Error("expected Supports=true for Java file with io.temporal import")
	}
	if a.Supports("file:///tmp/Workflow.java", notTemporal) {
		t.Error("expected Supports=false for Java file without io.temporal import")
	}
	if a.Supports("file:///tmp/workflow.py", temporal) {
		t.Error("expected Supports=false for non-Java file")
	}
}

func TestJavaDeterminismAnalyzerDetectsBannedCalls(t *testing.T) {
	a := NewJavaDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "java", "BadWorkflow.java")
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

	expected := []string{
		"temporal/no-time-now",
		"temporal/no-sleep",
		"temporal/no-random",
		"temporal/no-goroutine",
		"temporal/no-mutex",
	}
	for _, rule := range expected {
		if ruleCount[rule] == 0 {
			t.Errorf("expected at least one %s violation", rule)
		}
	}
}

func TestJavaDeterminismAnalyzerAcceptsGoodWorkflow(t *testing.T) {
	a := NewJavaDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "java", "GoodWorkflow.java")
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
