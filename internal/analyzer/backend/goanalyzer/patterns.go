package goanalyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

type PatternAnalyzer struct{}

func (a *PatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".go") {
		return false
	}
	s := string(content)
	return strings.Contains(s, rules.GoSDKImport) || strings.Contains(s, rules.GoSDKActivity)
}

func (a *PatternAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, uri, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var violations []rules.Violation

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Type.Params == nil {
			continue
		}

		scope := classifyFunc(fn)
		if scope == scopeWorkflow {
			violations = append(violations, checkWorkflowPatterns(fset, fn)...)
		}
		if scope == scopeActivity {
			violations = append(violations, checkActivityPatterns(fset, fn)...)
		}
	}

	return violations, nil
}

func checkWorkflowPatterns(fset *token.FileSet, fn *ast.FuncDecl) []rules.Violation {
	var violations []rules.Violation

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		callText := exprToString(call.Fun)

		if callText == "context.Background" || callText == "context.TODO" {
			violations = append(violations, rules.ContextPropagation.
				WithMessage("Use the workflow context instead of context.Background()/context.TODO() in workflows").
				At(posToRange(fset.Position(call.Pos()))))
		}

		if callText == "workflow.ExecuteActivity" {
			if !hasActivityOptionsInScope(fn.Body) {
				violations = append(violations, rules.ActivityTimeout.
					WithMessage("Set StartToCloseTimeout or ScheduleToCloseTimeout in ActivityOptions before calling ExecuteActivity").
					At(posToRange(fset.Position(call.Pos()))))
			}
		}

		return true
	})

	violations = append(violations, checkUnboundedLoops(fset, fn)...)

	return violations
}

func checkActivityPatterns(fset *token.FileSet, fn *ast.FuncDecl) []rules.Violation {
	var violations []rules.Violation

	if fn.Body == nil {
		return nil
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, result := range ret.Results {
			if isNakedErrorExpr(result) {
				violations = append(violations, rules.NakedError.
					WithMessage("Wrap activity errors with temporal.NewApplicationError for proper retry semantics").
					At(posToRange(fset.Position(ret.Pos()))))
				break
			}
		}
		return true
	})

	return violations
}

func hasActivityOptionsInScope(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		comp, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		typeName := exprToString(comp.Type)
		if typeName == "workflow.ActivityOptions" {
			for _, elt := range comp.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key := exprToString(kv.Key)
				if key == "StartToCloseTimeout" || key == "ScheduleToCloseTimeout" {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

func checkUnboundedLoops(fset *token.FileSet, fn *ast.FuncDecl) []rules.Violation {
	var violations []rules.Violation

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		forStmt, ok := n.(*ast.ForStmt)
		if !ok {
			return true
		}
		if forStmt.Init != nil || forStmt.Cond != nil || forStmt.Post != nil {
			return true
		}
		if hasContinueAsNew(fn.Body) {
			return true
		}
		violations = append(violations, rules.Unbounded.
			WithMessage("Infinite loop without workflow.NewContinueAsNewError risks history growth; add ContinueAsNew").
			At(posToRange(fset.Position(forStmt.Pos()))))
		return true
	})

	return violations
}

func hasContinueAsNew(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if exprToString(call.Fun) == "workflow.NewContinueAsNewError" {
			found = true
		}
		return !found
	})
	return found
}

func isNakedErrorExpr(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	funcName := exprToString(call.Fun)
	return funcName == "fmt.Errorf" || funcName == "errors.New"
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			return ident.Name + "." + e.Sel.Name
		}
	case *ast.Ident:
		return e.Name
	}
	return ""
}

func posToRange(pos token.Position) rules.Range {
	return rules.Range{
		StartLine: pos.Line - 1,
		StartCol:  pos.Column - 1,
		EndLine:   pos.Line - 1,
		EndCol:    pos.Column - 1,
	}
}
