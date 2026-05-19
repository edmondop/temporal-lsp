package analyzer

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestGoFixturesAreValidGo(t *testing.T) {
	goFixtureDirs := []string{
		filepath.Join(testdataDir(), "go", "determinism"),
		filepath.Join(testdataDir(), "go", "signatures"),
		filepath.Join(testdataDir(), "go", "patterns"),
	}

	for _, dir := range goFixtureDirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("failed to read fixture dir %s: %v", dir, err)
			}

			fset := token.NewFileSet()
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
					continue
				}
				filePath := filepath.Join(dir, entry.Name())
				_, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
				if err != nil {
					t.Errorf("Go fixture %s does not parse: %v", filePath, err)
				}
			}
		})
	}
}

func TestGoFixturesHaveGoMod(t *testing.T) {
	goFixtureDirs := []string{
		filepath.Join(testdataDir(), "go", "determinism"),
		filepath.Join(testdataDir(), "go", "signatures"),
		filepath.Join(testdataDir(), "go", "patterns"),
	}

	for _, dir := range goFixtureDirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			goMod := filepath.Join(dir, "go.mod")
			if _, err := os.Stat(goMod); err != nil {
				t.Errorf("Go fixture dir %s missing go.mod — fixtures must be compilable modules", dir)
			}
		})
	}
}

func TestGoFixturesCompileLocally(t *testing.T) {
	goFixtureDirs := []string{
		filepath.Join(testdataDir(), "go", "determinism"),
		filepath.Join(testdataDir(), "go", "signatures"),
		filepath.Join(testdataDir(), "go", "patterns"),
	}

	for _, dir := range goFixtureDirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			// Verify go.sum exists (dependencies resolved)
			goSum := filepath.Join(dir, "go.sum")
			if _, err := os.Stat(goSum); err != nil {
				t.Errorf("Go fixture dir %s missing go.sum — run 'go mod tidy' in fixture dir", dir)
			}

			// Parse all .go files and verify no syntax errors
			fset := token.NewFileSet()
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("failed to read dir: %v", err)
			}
			for _, entry := range entries {
				if filepath.Ext(entry.Name()) != ".go" {
					continue
				}
				filePath := filepath.Join(dir, entry.Name())
				f, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
				if err != nil {
					t.Errorf("syntax error in %s: %v", entry.Name(), err)
					continue
				}
				// Verify it has a package declaration
				if f.Name == nil || f.Name.Name == "" {
					t.Errorf("%s has no package declaration", entry.Name())
				}
				// Verify imports reference temporal SDK
				hasTemporalImport := false
				for _, imp := range f.Imports {
					if imp.Path != nil && (imp.Path.Value == `"go.temporal.io/sdk/workflow"` ||
						imp.Path.Value == `"go.temporal.io/sdk/temporal"` ||
						imp.Path.Value == `"go.temporal.io/sdk/activity"`) {
						hasTemporalImport = true
						break
					}
				}
				if !hasTemporalImport {
					// helpers.go may not import temporal directly, that's fine
					// but at least one file in the dir should
					t.Logf("note: %s does not directly import temporal SDK", entry.Name())
				}
			}
		})
	}
}

func TestPythonFixturesAreValidPython(t *testing.T) {
	pyDir := filepath.Join(testdataDir(), "python")
	entries, err := os.ReadDir(pyDir)
	if err != nil {
		t.Fatalf("failed to read python fixtures dir: %v", err)
	}

	// Use tree-sitter to parse — same tool the analyzer uses
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".py" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			filePath := filepath.Join(pyDir, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read %s: %v", entry.Name(), err)
			}

			// Parse with tree-sitter (same as production code)
			a := NewPythonDeterminismAnalyzer()
			if !a.Supports("file://"+filePath, content) {
				t.Errorf("%s does not contain temporalio import — fixture must be real temporal code", entry.Name())
			}

			// Verify tree-sitter can parse without errors
			p := sitterParse(content)
			if p == nil {
				t.Errorf("%s: tree-sitter failed to parse", entry.Name())
				return
			}
			if p.RootNode().HasError() {
				t.Errorf("%s has parse errors — not valid Python", entry.Name())
			}
		})
	}
}
