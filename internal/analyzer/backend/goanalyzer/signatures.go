package goanalyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

type SignatureAnalyzer struct{}

func (a *SignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".go") {
		return false
	}
	s := string(content)
	return strings.Contains(s, rules.GoSDKImport) || strings.Contains(s, rules.GoSDKActivity)
}

func (a *SignatureAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
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
		if scope == scopeOther {
			continue
		}

		violations = append(violations, checkSignature(fset, fn)...)
	}

	return violations, nil
}

type funcScope int

const (
	scopeOther funcScope = iota
	scopeWorkflow
	scopeActivity
)

func classifyFunc(fn *ast.FuncDecl) funcScope {
	params := fn.Type.Params.List
	if len(params) == 0 {
		return scopeOther
	}

	firstParamType := typeString(params[0].Type)
	switch firstParamType {
	case "workflow.Context":
		return scopeWorkflow
	case "context.Context":
		return scopeActivity
	}
	return scopeOther
}

func checkSignature(fset *token.FileSet, fn *ast.FuncDecl) []rules.Violation {
	var violations []rules.Violation

	nonCtxParams := fn.Type.Params.List[1:]

	rng := posToRange(fset.Position(fn.Name.Pos()))

	if len(nonCtxParams) > 1 {
		violations = append(violations, rules.SinglePayloadRule.
			WithMessage("Workflow/activity functions should accept a single struct parameter for forwards compatibility").
			At(rng))
	}

	primitiveCount := 0
	for _, p := range nonCtxParams {
		if isPrimitiveType(p.Type) {
			primitiveCount++
		}
	}
	if primitiveCount >= 2 {
		violations = append(violations, rules.PrimitiveParamsRule.
			WithMessage("Use a struct instead of multiple primitive parameters for workflow/activity inputs").
			At(rng))
	}

	if fn.Type.Results != nil {
		returnCount := 0
		for _, field := range fn.Type.Results.List {
			if len(field.Names) == 0 {
				returnCount++
			} else {
				returnCount += len(field.Names)
			}
		}
		if returnCount > 2 {
			violations = append(violations, rules.SingleReturnRule.
				WithMessage("Workflow/activity functions should return at most (result, error) — wrap multiple values in a struct").
				At(rng))
		}
	}

	return violations
}

func isPrimitiveType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string", "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "bool", "byte", "rune":
			return true
		}
	case *ast.ArrayType:
		if ident, ok := t.Elt.(*ast.Ident); ok && ident.Name == "byte" && t.Len == nil {
			return true
		}
	}
	return false
}

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}
