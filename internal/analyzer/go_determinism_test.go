package analyzer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func TestGoDeterminismAnalyzerSupportsTemporalWorkflowFiles(t *testing.T) {
	a := NewGoDeterminismAnalyzer()

	goWithTemporal := []byte(`package foo
import "go.temporal.io/sdk/workflow"
func F(ctx workflow.Context) {}
`)
	goWithoutTemporal := []byte(`package foo
import "fmt"
func F() {}
`)

	if !a.Supports("file:///tmp/workflow.go", goWithTemporal) {
		t.Error("expected Supports to return true for Go file with Temporal import")
	}
	if a.Supports("file:///tmp/main.go", goWithoutTemporal) {
		t.Error("expected Supports to return false for Go file without Temporal import")
	}
	if a.Supports("file:///tmp/workflow.py", goWithTemporal) {
		t.Error("expected Supports to return false for non-Go file")
	}
}

func TestGoDeterminismAnalyzerDetectsTimeNow(t *testing.T) {
	a := NewGoDeterminismAnalyzer()

	dir := filepath.Join(testdataDir(), "go", "determinism")
	uri := "file://" + filepath.Join(dir, "workflow.go")
	content, err := os.ReadFile(filepath.Join(dir, "workflow.go"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	violations, err := a.Analyze(uri, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(violations) == 0 {
		t.Fatal("expected at least one violation for time.Now() in workflow")
	}

	found := false
	for _, v := range violations {
		if v.RuleID == "temporal/non-deterministic" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected violation with RuleID 'temporal/non-deterministic', got: %+v", violations)
	}
}

func TestGoDeterminismAnalyzerDetectsTransitiveNonDeterminism(t *testing.T) {
	a := NewGoDeterminismAnalyzer()

	dir := filepath.Join(testdataDir(), "go", "determinism")
	uri := "file://" + filepath.Join(dir, "workflow_transitive.go")
	content, err := os.ReadFile(filepath.Join(dir, "workflow_transitive.go"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	violations, err := a.Analyze(uri, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(violations) == 0 {
		t.Fatal("expected at least one violation for transitive non-determinism via getCurrentTime()")
	}
}
