package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoSignatureAnalyzerDetectsBadSignatures(t *testing.T) {
	a := NewGoSignatureAnalyzer()

	dir := filepath.Join(testdataDir(), "go", "signatures")
	filePath := filepath.Join(dir, "bad.go")
	uri := "file://" + filePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	violations, err := a.Analyze(uri, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	ruleCount := map[ID]int{}
	for _, v := range violations {
		ruleCount[v.RuleID]++
	}

	if ruleCount["temporal/primitive-params"] == 0 {
		t.Error("expected primitive-params violation for BadWorkflow with (name string, age int)")
	}

	if ruleCount["temporal/single-payload"] == 0 {
		t.Error("expected single-payload violation for BadWorkflow with >1 non-context params")
	}

	if ruleCount["temporal/single-return"] == 0 {
		t.Error("expected single-return violation for BadActivity returning (string, int, error)")
	}
}

func TestGoSignatureAnalyzerAcceptsGoodSignatures(t *testing.T) {
	a := NewGoSignatureAnalyzer()

	dir := filepath.Join(testdataDir(), "go", "signatures")
	filePath := filepath.Join(dir, "good.go")
	uri := "file://" + filePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	violations, err := a.Analyze(uri, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(violations) != 0 {
		t.Errorf("expected no violations for good signatures, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s: %s (line %d)", v.RuleID, v.Message, v.Range.StartLine+1)
		}
	}
}
