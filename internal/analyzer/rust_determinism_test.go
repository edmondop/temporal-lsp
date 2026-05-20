package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRustDeterminismAnalyzerSupportsTemporalFiles(t *testing.T) {
	a := NewRustDeterminismAnalyzer()

	temporal := []byte(`use temporal_sdk::prelude::*;`)
	notTemporal := []byte(`use std::io;`)

	if !a.Supports("file:///tmp/workflow.rs", temporal) {
		t.Error("expected Supports=true for Rust file with temporal_sdk import")
	}
	if a.Supports("file:///tmp/workflow.rs", notTemporal) {
		t.Error("expected Supports=false for Rust file without temporal_sdk import")
	}
	if a.Supports("file:///tmp/workflow.py", temporal) {
		t.Error("expected Supports=false for non-Rust file")
	}
}

func TestRustDeterminismAnalyzerDetectsBannedCalls(t *testing.T) {
	a := NewRustDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "rust", "bad_workflow.rs")
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
		"temporal/no-io",
		"temporal/no-goroutine",
		"temporal/no-mutex",
	}
	for _, rule := range expected {
		if ruleCount[rule] == 0 {
			t.Errorf("expected at least one %s violation", rule)
		}
	}
}

func TestRustDeterminismAnalyzerAcceptsGoodWorkflow(t *testing.T) {
	a := NewRustDeterminismAnalyzer()

	filePath := filepath.Join(testdataDir(), "rust", "good_workflow.rs")
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
