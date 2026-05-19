package analyzer

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"go.temporal.io/sdk/contrib/tools/workflowcheck/workflow"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

type GoDeterminismAnalyzer struct{}

func NewGoDeterminismAnalyzer() *GoDeterminismAnalyzer {
	return &GoDeterminismAnalyzer{}
}

func (a *GoDeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".go") {
		return false
	}
	return strings.Contains(string(content), `"go.temporal.io/sdk/workflow"`)
}

func (a *GoDeterminismAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	dir := findModuleDir(uriToPath(uri))
	if dir == "" {
		return nil, nil
	}

	pkgs, err := loadPackages(dir)
	if err != nil {
		return nil, err
	}

	wfChecker := workflow.NewChecker(workflow.Config{})
	analyzers := []*analysis.Analyzer{wfChecker.NewAnalyzer()}

	graph, err := checker.Analyze(analyzers, pkgs, nil)
	if err != nil {
		return nil, err
	}

	filePath := uriToPath(uri)
	var violations []Violation

	for act := range graph.All() {
		if act.Err != nil {
			continue
		}
		for _, diag := range act.Diagnostics {
			pos := act.Package.Fset.Position(diag.Pos)
			if !samePath(pos.Filename, filePath) {
				continue
			}
			endPos := pos
			if diag.End != token.NoPos {
				endPos = act.Package.Fset.Position(diag.End)
			}
			violations = append(violations, Violation{
				RuleID:   "temporal/non-deterministic",
				Message:  diag.Message,
				Severity: 1,
				Range: Range{
					StartLine: pos.Line - 1,
					StartCol:  pos.Column - 1,
					EndLine:   endPos.Line - 1,
					EndCol:    endPos.Column - 1,
				},
			})
		}
	}

	return violations, nil
}

func loadPackages(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  dir,
	}
	return packages.Load(cfg, "./...")
}

func uriToPath(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}

func findModuleDir(path string) string {
	dir := filepath.Dir(path)
	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func samePath(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return a == b
	}
	return absA == absB
}
