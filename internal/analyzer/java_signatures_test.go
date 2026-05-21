package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJavaSignatureAnalyzerDetectsBadSignatures(t *testing.T) {
	a := NewJavaSignatureAnalyzer()

	filePath := filepath.Join(testdataDir(), "java", "BadSignatures.java")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	violations, err := a.Analyze("file://"+filePath, content)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	ruleCount := map[ID]int{}
	for _, v := range violations {
		ruleCount[v.RuleID]++
	}

	if ruleCount["temporal/single-payload"] < 2 {
		t.Errorf("expected at least 2 single-payload violations, got %d", ruleCount["temporal/single-payload"])
	}
	if ruleCount["temporal/primitive-params"] < 2 {
		t.Errorf("expected at least 2 primitive-params violations, got %d", ruleCount["temporal/primitive-params"])
	}
}

func TestJavaSignatureAnalyzerAcceptsGoodSignatures(t *testing.T) {
	a := NewJavaSignatureAnalyzer()

	filePath := filepath.Join(testdataDir(), "java", "GoodSignatures.java")
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
