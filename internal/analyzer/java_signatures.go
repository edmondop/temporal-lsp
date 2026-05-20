package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

var javaPrimitiveTypes = map[string]bool{
	"int":     true,
	"long":    true,
	"double":  true,
	"float":   true,
	"boolean": true,
	"String":  true,
	"byte":    true,
	"short":   true,
	"char":    true,
	"Integer": true,
	"Long":    true,
	"Double":  true,
	"Float":   true,
	"Boolean": true,
}

type JavaSignatureAnalyzer struct{}

func NewJavaSignatureAnalyzer() *JavaSignatureAnalyzer {
	return &JavaSignatureAnalyzer{}
}

func (a *JavaSignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".java") {
		return false
	}
	return strings.Contains(string(content), "io.temporal")
}

func (a *JavaSignatureAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()
	var violations []Violation

	walkNode(root, func(n *sitter.Node) {
		if n.Type() != "method_declaration" {
			return
		}
		if !hasJavaAnnotation(n, content, []string{"WorkflowMethod", "ActivityMethod"}) {
			return
		}
		violations = append(violations, checkJavaSignature(n, content)...)
	})

	return violations, nil
}

func checkJavaSignature(method *sitter.Node, content []byte) []Violation {
	var violations []Violation

	params := findChildOfType(method, "formal_parameters")
	if params == nil {
		return nil
	}

	var paramNodes []*sitter.Node
	for i := 0; i < int(params.ChildCount()); i++ {
		child := params.Child(i)
		if child.Type() == "formal_parameter" || child.Type() == "spread_parameter" {
			paramNodes = append(paramNodes, child)
		}
	}

	const ref = "https://github.com/jlegrone/100-temporal-mistakes/blob/main/src/using_more_than_one_input_response_payload/"

	nameNode := findJavaMethodName(method)
	startLine := int(method.StartPoint().Row)
	startCol := int(method.StartPoint().Column)
	if nameNode != nil {
		startLine = int(nameNode.StartPoint().Row)
		startCol = int(nameNode.StartPoint().Column)
	}

	if len(paramNodes) > 1 {
		violations = append(violations, Violation{
			RuleID:    "temporal/single-payload",
			Message:   "Workflow/activity methods should accept a single parameter object (use a POJO for multiple values)",
			Severity:  2,
			Range:     Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol},
			Reference: ref,
		})
	}

	primitiveCount := 0
	for _, p := range paramNodes {
		if isJavaPrimitiveParam(p, content) {
			primitiveCount++
		}
	}
	if primitiveCount >= 2 {
		violations = append(violations, Violation{
			RuleID:    "temporal/primitive-params",
			Message:   "Use a POJO instead of multiple primitive parameters for workflow/activity inputs",
			Severity:  2,
			Range:     Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol},
			Reference: ref,
		})
	}

	return violations
}

func isJavaPrimitiveParam(param *sitter.Node, content []byte) bool {
	for i := 0; i < int(param.ChildCount()); i++ {
		child := param.Child(i)
		// type nodes in Java tree-sitter: type_identifier, integral_type, floating_point_type, boolean_type
		switch child.Type() {
		case "type_identifier":
			return javaPrimitiveTypes[child.Content(content)]
		case "integral_type", "floating_point_type", "boolean_type":
			return true
		}
	}
	return false
}

func findJavaMethodName(method *sitter.Node) *sitter.Node {
	for i := 0; i < int(method.ChildCount()); i++ {
		child := method.Child(i)
		if child.Type() == "identifier" {
			return child
		}
	}
	return nil
}
