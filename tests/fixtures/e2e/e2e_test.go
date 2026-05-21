package e2e

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/edmondop/temporal-lsp/internal/analyzer"
	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var update = flag.Bool("update", false, "update golden expected files")

type diagnostic struct {
	Rule     rules.ID `json:"rule"`
	Message  string   `json:"message"`
	Severity int      `json:"severity"`
	Range    Range    `json:"range"`
}

type Range struct {
	StartLine int `json:"startLine"`
	StartCol  int `json:"startCol"`
	EndLine   int `json:"endLine"`
	EndCol    int `json:"endCol"`
}

func TestE2E(t *testing.T) {
	analyzers := analyzer.AllAnalyzers()
	root := fixtureRoot()

	languages := []struct {
		dir  string
		glob string
	}{
		{"python", "*.py"},
		{"java", "*.java"},
		{"rust", "*.rs"},
	}

	for _, lang := range languages {
		t.Run(lang.dir, func(t *testing.T) {
			dir := filepath.Join(root, lang.dir)
			matches, err := filepath.Glob(filepath.Join(dir, lang.glob))
			if err != nil {
				t.Fatalf("glob failed: %v", err)
			}
			for _, fixture := range matches {
				name := strings.TrimSuffix(filepath.Base(fixture), filepath.Ext(fixture))
				t.Run(name, func(t *testing.T) {
					runFixture(t, analyzers, fixture, filepath.Join(root, "expected", lang.dir, name+".json"))
				})
			}
		})
	}

	t.Run("go", func(t *testing.T) {
		goDir := filepath.Join(root, "go")
		entries, err := os.ReadDir(goDir)
		if err != nil {
			t.Fatalf("failed to read go fixture dir: %v", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			t.Run(name, func(t *testing.T) {
				workflowFile := filepath.Join(goDir, name, "workflow.go")
				runFixture(t, analyzers, workflowFile, filepath.Join(root, "expected", "go", name+".json"))
			})
		}
	})
}

func runFixture(t *testing.T, analyzers []rules.Analyzer, fixture string, expectedFile string) {
	t.Helper()

	content, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", fixture, err)
	}

	uri := "file://" + fixture

	var diagnostics []diagnostic
	for _, a := range analyzers {
		if !a.Supports(uri, content) {
			continue
		}
		violations, err := a.Analyze(uri, content)
		if err != nil {
			t.Fatalf("analyzer error: %v", err)
		}
		for _, v := range violations {
			diagnostics = append(diagnostics, diagnostic{
				Rule:     v.RuleID,
				Message:  v.Message,
				Severity: v.Severity,
				Range: Range{
					StartLine: v.Range.StartLine,
					StartCol:  v.Range.StartCol,
					EndLine:   v.Range.EndLine,
					EndCol:    v.Range.EndCol,
				},
			})
		}
	}

	if diagnostics == nil {
		diagnostics = []diagnostic{}
	}

	actual, err := json.MarshalIndent(diagnostics, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal diagnostics: %v", err)
	}
	actual = append(actual, '\n')

	if *update {
		dir := filepath.Dir(expectedFile)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create expected dir: %v", err)
		}
		if err := os.WriteFile(expectedFile, actual, 0o644); err != nil {
			t.Fatalf("failed to write expected file: %v", err)
		}
		return
	}

	expected, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("missing expected file %s (run with -update to create)", expectedFile)
	}

	if string(actual) != string(expected) {
		t.Errorf("diagnostics mismatch for %s\n\ngot:\n%s\nwant:\n%s", filepath.Base(fixture), actual, expected)
	}
}

var expectedRulesByLanguage = map[string][]rules.ID{
	"python": {
		rules.NoTimeNow,
		rules.NoSleep,
		rules.NoRandom,
		rules.NoIO,
		rules.NoGoroutine,
		rules.NoMutex,
		rules.NoChannel,
		rules.NoEnvAccess,
		rules.NoStandardLogging,
		rules.UnboundedLoop,
		rules.SinglePayload,
		rules.PrimitiveParams,
	},
	"java": {
		rules.NoTimeNow,
		rules.NoSleep,
		rules.NoRandom,
		rules.NoIO,
		rules.NoGoroutine,
		rules.NoMutex,
		rules.UnboundedLoop,
		rules.SinglePayload,
		rules.PrimitiveParams,
	},
	"rust": {
		rules.NoTimeNow,
		rules.NoSleep,
		rules.NoRandom,
		rules.NoIO,
		rules.NoGoroutine,
		rules.NoMutex,
		rules.UnboundedLoop,
		rules.SinglePayload,
		rules.PrimitiveParams,
	},
	"go": {
		rules.NonDeterministic,
		rules.NoContextPropagation,
		rules.UnboundedLoop,
		rules.SinglePayload,
		rules.PrimitiveParams,
		rules.SingleReturn,
	},
}

func TestCoverageAllRulesHaveFixtures(t *testing.T) {
	root := fixtureRoot()
	analyzers := analyzer.AllAnalyzers()

	for lang, expectedRules := range expectedRulesByLanguage {
		t.Run(lang, func(t *testing.T) {
			var isGo bool
			var glob string
			switch lang {
			case "go":
				isGo = true
			case "python":
				glob = "*.py"
			case "java":
				glob = "*.java"
			case "rust":
				glob = "*.rs"
			}

			covered := collectCoveredRules(t, root, lang, glob, isGo, analyzers)

			for _, rule := range expectedRules {
				if !covered[rule] {
					t.Errorf("rule %s has no e2e fixture triggering it — add a fixture to tests/fixtures/e2e/%s/", rule, lang)
				}
			}
		})
	}
}

func collectCoveredRules(t *testing.T, root, langDir, glob string, isGo bool, analyzers []rules.Analyzer) map[rules.ID]bool {
	t.Helper()
	covered := map[rules.ID]bool{}

	fixtures := listFixtures(t, root, langDir, glob, isGo)
	for _, fixture := range fixtures {
		content, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatalf("failed to read %s: %v", fixture, err)
		}
		uri := "file://" + fixture
		for _, a := range analyzers {
			if !a.Supports(uri, content) {
				continue
			}
			violations, err := a.Analyze(uri, content)
			if err != nil {
				continue
			}
			for _, v := range violations {
				covered[v.RuleID] = true
			}
		}
	}
	return covered
}

func listFixtures(t *testing.T, root, langDir, glob string, isGo bool) []string {
	t.Helper()
	if isGo {
		goDir := filepath.Join(root, langDir)
		entries, err := os.ReadDir(goDir)
		if err != nil {
			t.Fatalf("failed to read %s: %v", goDir, err)
		}
		var fixtures []string
		for _, entry := range entries {
			if entry.IsDir() {
				fixtures = append(fixtures, filepath.Join(goDir, entry.Name(), "workflow.go"))
			}
		}
		return fixtures
	}

	dir := filepath.Join(root, langDir)
	matches, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	return matches
}

func fixtureRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}
