package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type GoPatternAnalyzer struct{}

func NewGoPatternAnalyzer() *GoPatternAnalyzer {
	return &GoPatternAnalyzer{}
}

func (a *GoPatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".go") {
		return false
	}
	s := string(content)
	return strings.Contains(s, `"go.temporal.io/sdk/workflow"`) ||
		strings.Contains(s, `"go.temporal.io/sdk/activity"`)
}

func (a *GoPatternAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, uri, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var violations []Violation

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Type.Params == nil {
			continue
		}

		scope := classifyFunc(fn)
		if scope == scopeOther {
			continue
		}

		if scope == scopeWorkflow {
			violations = append(violations, checkWorkflowPatterns(fset, fn)...)
		}
		if scope == scopeActivity {
			violations = append(violations, checkActivityPatterns(fset, fn)...)
		}
	}

	return violations, nil
}

func checkWorkflowPatterns(fset *token.FileSet, fn *ast.FuncDecl) []Violation {
	var violations []Violation

	const ref = "https://github.com/jlegrone/100-temporal-mistakes"

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		callText := exprToString(call.Fun)

		// no-context-propagation: context.Background() or context.TODO() in workflow
		if callText == "context.Background" || callText == "context.TODO" {
			pos := fset.Position(call.Pos())
			violations = append(violations, Violation{
				RuleID:    "temporal/no-context-propagation",
				Message:   "Use the workflow context instead of context.Background()/context.TODO() in workflows",
				Severity:  1,
				Range:     posToRange(pos),
				Reference: ref,
			})
		}

		// activity-timeout-required: workflow.ExecuteActivity without ActivityOptions
		if callText == "workflow.ExecuteActivity" {
			if !hasActivityOptionsInScope(fn.Body) {
				pos := fset.Position(call.Pos())
				violations = append(violations, Violation{
					RuleID:    "temporal/activity-timeout-required",
					Message:   "Set StartToCloseTimeout or ScheduleToCloseTimeout in ActivityOptions before calling ExecuteActivity",
					Severity:  2,
					Range:     posToRange(pos),
					Reference: ref,
				})
			}
		}

		return true
	})

	// unbounded-loop: for {} without ContinueAsNew
	violations = append(violations, checkUnboundedLoops(fset, fn, ref)...)

	return violations
}

func checkActivityPatterns(fset *token.FileSet, fn *ast.FuncDecl) []Violation {
	var violations []Violation
	const ref = "https://github.com/jlegrone/100-temporal-mistakes"

	if fn.Body == nil {
		return nil
	}

	// no-naked-error: check return statements for bare errors
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		for _, result := range ret.Results {
			if isNakedErrorExpr(result) {
				pos := fset.Position(ret.Pos())
				violations = append(violations, Violation{
					RuleID:    "temporal/no-naked-error",
					Message:   "Wrap activity errors with temporal.NewApplicationError for proper retry semantics",
					Severity:  2,
					Range:     posToRange(pos),
					Reference: ref,
				})
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

func checkUnboundedLoops(fset *token.FileSet, fn *ast.FuncDecl, ref string) []Violation {
	var violations []Violation

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		forStmt, ok := n.(*ast.ForStmt)
		if !ok {
			return true
		}
		// Only flag infinite loops (no init, no cond, no post)
		if forStmt.Init != nil || forStmt.Cond != nil || forStmt.Post != nil {
			return true
		}
		// Check if ContinueAsNew is called anywhere in the function
		if hasContinueAsNew(fn.Body) {
			return true
		}
		pos := fset.Position(forStmt.Pos())
		violations = append(violations, Violation{
			RuleID:    "temporal/unbounded-loop",
			Message:   "Infinite loop without workflow.NewContinueAsNewError risks history growth; add ContinueAsNew",
			Severity:  2,
			Range:     posToRange(pos),
			Reference: ref,
		})
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
