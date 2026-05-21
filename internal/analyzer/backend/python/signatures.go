package python

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var primitiveTypes = map[string]bool{
	"str":   true,
	"int":   true,
	"float": true,
	"bool":  true,
	"bytes": true,
}

type SignatureAnalyzer struct{}

func (a *SignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".py") {
		return false
	}
	s := string(content)
	return strings.Contains(s, rules.PythonSDKImportFrom) || strings.Contains(s, rules.PythonSDKImport)
}

func (a *SignatureAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()
	var violations []rules.Violation

	for _, fn := range findWorkflowRunFunctions(root, content) {
		violations = append(violations, checkSignature(fn, content)...)
	}
	for _, fn := range findActivityFunctions(root, content) {
		violations = append(violations, checkSignature(fn, content)...)
	}

	return violations, nil
}

func findWorkflowRunFunctions(root *sitter.Node, content []byte) []*sitter.Node {
	var funcs []*sitter.Node

	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() != rules.NodeDecoratedDef {
			continue
		}
		if !hasAnyDecorator(child, content, workflowDefnDecorators) {
			continue
		}
		classDef := rules.FindChildOfType(child, rules.NodeClassDef)
		if classDef == nil {
			continue
		}
		body := rules.FindChildOfType(classDef, rules.NodeBlock)
		if body == nil {
			continue
		}
		for j := 0; j < int(body.ChildCount()); j++ {
			method := body.Child(j)
			if method.Type() != rules.NodeDecoratedDef {
				continue
			}
			if !hasAnyDecorator(method, content, workflowRunDecorators) {
				continue
			}
			funcDef := rules.FindChildOfType(method, rules.NodeFunctionDef)
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
		if child.Type() != rules.NodeDecoratedDef {
			continue
		}
		if !hasActivityDecorator(child, content) {
			continue
		}
		funcDef := rules.FindChildOfType(child, rules.NodeFunctionDef)
		if funcDef != nil {
			funcs = append(funcs, funcDef)
		}
	}

	return funcs
}

func hasActivityDecorator(decoratedDef *sitter.Node, content []byte) bool {
	for i := 0; i < int(decoratedDef.ChildCount()); i++ {
		child := decoratedDef.Child(i)
		if child.Type() != rules.NodeDecorator {
			continue
		}
		text := child.Content(content)
		if strings.Contains(text, "activity.defn") {
			return true
		}
	}
	return false
}

func checkSignature(funcDef *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation

	params := rules.FindChildOfType(funcDef, rules.NodeParameters)
	if params == nil {
		return nil
	}

	nonSelfParams := getNonSelfParams(params, content)

	nameNode := findFuncName(funcDef)
	startLine := int(funcDef.StartPoint().Row)
	startCol := int(funcDef.StartPoint().Column)
	if nameNode != nil {
		startLine = int(nameNode.StartPoint().Row)
		startCol = int(nameNode.StartPoint().Column)
	}
	rng := rules.Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol}

	if len(nonSelfParams) > 1 {
		violations = append(violations, rules.SinglePayloadRule.
			WithMessage("Workflow/activity functions should accept a single parameter (use a dataclass for multiple values)").
			At(rng))
	}

	primitiveCount := 0
	for _, p := range nonSelfParams {
		if isPrimitiveParam(p, content) {
			primitiveCount++
		}
	}
	if primitiveCount >= 2 {
		violations = append(violations, rules.PrimitiveParamsRule.
			WithMessage("Use a dataclass instead of multiple primitive parameters for workflow/activity inputs").
			At(rng))
	}

	return violations
}

func getNonSelfParams(params *sitter.Node, content []byte) []*sitter.Node {
	var result []*sitter.Node
	for i := 0; i < int(params.ChildCount()); i++ {
		child := params.Child(i)
		switch child.Type() {
		case rules.NodeIdentifier, "typed_parameter", "default_parameter", "typed_default_parameter":
		default:
			continue
		}
		if getParamName(child, content) == "self" {
			continue
		}
		result = append(result, child)
	}
	return result
}

func getParamName(param *sitter.Node, content []byte) string {
	switch param.Type() {
	case rules.NodeIdentifier:
		return param.Content(content)
	case "typed_parameter", "default_parameter", "typed_default_parameter":
		if param.ChildCount() > 0 {
			return param.Child(0).Content(content)
		}
	}
	return ""
}

func isPrimitiveParam(param *sitter.Node, content []byte) bool {
	if param.Type() != "typed_parameter" && param.Type() != "typed_default_parameter" {
		return false
	}
	for i := 0; i < int(param.ChildCount()); i++ {
		child := param.Child(i)
		if child.Type() == "type" {
			typeText := strings.TrimSpace(child.Content(content))
			return primitiveTypes[typeText]
		}
	}
	return false
}

func findFuncName(funcDef *sitter.Node) *sitter.Node {
	for i := 0; i < int(funcDef.ChildCount()); i++ {
		child := funcDef.Child(i)
		if child.Type() == rules.NodeIdentifier {
			return child
		}
	}
	return nil
}
