package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

var pythonPrimitiveTypes = map[string]bool{
	"str":   true,
	"int":   true,
	"float": true,
	"bool":  true,
	"bytes": true,
}

type PythonSignatureAnalyzer struct{}

func NewPythonSignatureAnalyzer() *PythonSignatureAnalyzer {
	return &PythonSignatureAnalyzer{}
}

func (a *PythonSignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".py") {
		return false
	}
	s := string(content)
	return strings.Contains(s, "from temporalio") || strings.Contains(s, "import temporalio")
}

func (a *PythonSignatureAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()

	var violations []Violation

	// Check @workflow.run methods
	for _, fn := range findWorkflowRunFunctions(root, content) {
		violations = append(violations, checkPythonSignature(fn, content)...)
	}

	// Check @activity.defn functions
	for _, fn := range findActivityFunctions(root, content) {
		violations = append(violations, checkPythonSignature(fn, content)...)
	}

	return violations, nil
}

func findWorkflowRunFunctions(root *sitter.Node, content []byte) []*sitter.Node {
	var funcs []*sitter.Node

	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() != "decorated_definition" {
			continue
		}
		if !hasAnyDecorator(child, content, workflowDefnDecorators) {
			continue
		}
		classDef := findChildOfType(child, "class_definition")
		if classDef == nil {
			continue
		}
		body := findChildOfType(classDef, "block")
		if body == nil {
			continue
		}
		for j := 0; j < int(body.ChildCount()); j++ {
			method := body.Child(j)
			if method.Type() != "decorated_definition" {
				continue
			}
			if !hasAnyDecorator(method, content, workflowRunDecorators) {
				continue
			}
			funcDef := findChildOfType(method, "function_definition")
			if funcDef != nil {
				funcs = append(funcs, funcDef)
			}
		}
	}

	return funcs
}

func findActivityFunctions(root *sitter.Node, content []byte) []*sitter.Node {
	var funcs []*sitter.Node

	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() != "decorated_definition" {
			continue
		}
		if !hasDecorator(child, content, "activity.defn") {
			continue
		}
		funcDef := findChildOfType(child, "function_definition")
		if funcDef != nil {
			funcs = append(funcs, funcDef)
		}
	}

	return funcs
}

func checkPythonSignature(funcDef *sitter.Node, content []byte) []Violation {
	var violations []Violation

	params := findChildOfType(funcDef, "parameters")
	if params == nil {
		return nil
	}

	// Collect non-self parameters
	nonSelfParams := getNonSelfParams(params, content)

	const ref = "https://github.com/jlegrone/100-temporal-mistakes/blob/main/src/using_more_than_one_input_response_payload/"

	funcNameNode := findFuncName(funcDef)
	startLine := int(funcDef.StartPoint().Row)
	startCol := int(funcDef.StartPoint().Column)
	if funcNameNode != nil {
		startLine = int(funcNameNode.StartPoint().Row)
		startCol = int(funcNameNode.StartPoint().Column)
	}

	// single-payload: >1 non-self parameter
	if len(nonSelfParams) > 1 {
		violations = append(violations, Violation{
			RuleID:    "temporal/single-payload",
			Message:   "Workflow/activity functions should accept a single parameter (use a dataclass for multiple values)",
			Severity:  2,
			Range:     Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol},
			Reference: ref,
		})
	}

	// primitive-params: 2+ primitive type-annotated parameters
	primitiveCount := 0
	for _, p := range nonSelfParams {
		if isPythonPrimitiveParam(p, content) {
			primitiveCount++
		}
	}
	if primitiveCount >= 2 {
		violations = append(violations, Violation{
			RuleID:    "temporal/primitive-params",
			Message:   "Use a dataclass instead of multiple primitive parameters for workflow/activity inputs",
			Severity:  2,
			Range:     Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol},
			Reference: ref,
		})
	}

	return violations
}

func getNonSelfParams(params *sitter.Node, content []byte) []*sitter.Node {
	var result []*sitter.Node
	for i := 0; i < int(params.ChildCount()); i++ {
		child := params.Child(i)
		// Skip punctuation (parentheses, commas)
		if child.Type() != "identifier" && child.Type() != "typed_parameter" && child.Type() != "default_parameter" && child.Type() != "typed_default_parameter" {
			continue
		}
		// Skip "self"
		paramName := getParamName(child, content)
		if paramName == "self" {
			continue
		}
		result = append(result, child)
	}
	return result
}

func getParamName(param *sitter.Node, content []byte) string {
	switch param.Type() {
	case "identifier":
		return param.Content(content)
	case "typed_parameter":
		// First child is the name identifier
		if param.ChildCount() > 0 {
			return param.Child(0).Content(content)
		}
	case "default_parameter", "typed_default_parameter":
		if param.ChildCount() > 0 {
			return param.Child(0).Content(content)
		}
	}
	return ""
}

func isPythonPrimitiveParam(param *sitter.Node, content []byte) bool {
	if param.Type() != "typed_parameter" && param.Type() != "typed_default_parameter" {
		return false
	}
	// Find the type annotation (after the colon)
	for i := 0; i < int(param.ChildCount()); i++ {
		child := param.Child(i)
		if child.Type() == "type" {
			typeText := strings.TrimSpace(child.Content(content))
			return pythonPrimitiveTypes[typeText]
		}
	}
	return false
}

func findFuncName(funcDef *sitter.Node) *sitter.Node {
	for i := 0; i < int(funcDef.ChildCount()); i++ {
		child := funcDef.Child(i)
		if child.Type() == "identifier" {
			return child
		}
	}
	return nil
}
