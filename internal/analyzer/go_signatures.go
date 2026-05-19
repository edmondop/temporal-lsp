package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type GoSignatureAnalyzer struct{}

func NewGoSignatureAnalyzer() *GoSignatureAnalyzer {
	return &GoSignatureAnalyzer{}
}

func (a *GoSignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".go") {
		return false
	}
	s := string(content)
	return strings.Contains(s, `"go.temporal.io/sdk/workflow"`) ||
		strings.Contains(s, `"go.temporal.io/sdk/activity"`)
}

func (a *GoSignatureAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
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

		violations = append(violations, checkSignature(fset, fn, scope)...)
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

func checkSignature(fset *token.FileSet, fn *ast.FuncDecl, scope funcScope) []Violation {
	var violations []Violation

	nonCtxParams := getNonContextParams(fn.Type.Params.List)

	const ref = "https://github.com/jlegrone/100-temporal-mistakes/blob/main/src/using_more_than_one_input_response_payload/"

	// single-payload: >1 non-context parameter
	if len(nonCtxParams) > 1 {
		pos := fset.Position(fn.Name.Pos())
		violations = append(violations, Violation{
			RuleID:    "temporal/single-payload",
			Message:   "Workflow/activity functions should accept a single struct parameter for forwards compatibility",
			Severity:  2,
			Range:     posToRange(pos),
			Reference: ref,
		})
	}

	// primitive-params: 2+ primitive parameters
	primitiveCount := 0
	for _, p := range nonCtxParams {
		if isPrimitiveType(p.Type) {
			primitiveCount++
		}
	}
	if primitiveCount >= 2 {
		pos := fset.Position(fn.Name.Pos())
		violations = append(violations, Violation{
			RuleID:    "temporal/primitive-params",
			Message:   "Use a struct instead of multiple primitive parameters for workflow/activity inputs",
			Severity:  2,
			Range:     posToRange(pos),
			Reference: ref,
		})
	}

	// single-return: >2 return values (more than result + error)
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
			pos := fset.Position(fn.Name.Pos())
			violations = append(violations, Violation{
				RuleID:    "temporal/single-return",
				Message:   "Workflow/activity functions should return at most (result, error) — wrap multiple values in a struct",
				Severity:  2,
				Range:     posToRange(pos),
				Reference: ref,
			})
		}
	}

	return violations
}

func getNonContextParams(params []*ast.Field) []*ast.Field {
	if len(params) == 0 {
		return nil
	}
	// Skip the first param (context)
	return params[1:]
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
		// []byte is primitive
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

func posToRange(pos token.Position) Range {
	return Range{
		StartLine: pos.Line - 1,
		StartCol:  pos.Column - 1,
		EndLine:   pos.Line - 1,
		EndCol:    pos.Column - 1,
	}
}
