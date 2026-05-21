package java

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var primitiveTypes = map[string]bool{
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

var signatureAnnotations = []string{
	rules.JavaAnnotationWorkflowMethod,
	rules.JavaAnnotationActivityMethod,
}

type SignatureAnalyzer struct{}

func (a *SignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".java") {
		return false
	}
	return strings.Contains(string(content), rules.JavaSDKImport)
}

func (a *SignatureAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	var violations []rules.Violation
	rules.WalkNode(tree.RootNode(), func(n *sitter.Node) {
		if n.Type() != rules.NodeMethodDecl {
			return
		}
		if !hasAnnotation(n, content, signatureAnnotations) {
			return
		}
		violations = append(violations, checkSignature(n, content)...)
	})

	return violations, nil
}

func checkSignature(method *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation

	params := rules.FindChildOfType(method, rules.NodeFormalParameters)
	if params == nil {
		return nil
	}

	var paramNodes []*sitter.Node
	for i := 0; i < int(params.ChildCount()); i++ {
		child := params.Child(i)
		if child.Type() == rules.NodeFormalParameter || child.Type() == rules.NodeSpreadParameter {
			paramNodes = append(paramNodes, child)
		}
	}

	nameNode := findMethodName(method)
	startLine := int(method.StartPoint().Row)
	startCol := int(method.StartPoint().Column)
	if nameNode != nil {
		startLine = int(nameNode.StartPoint().Row)
		startCol = int(nameNode.StartPoint().Column)
	}
	rng := rules.Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol}

	if len(paramNodes) > 1 {
		violations = append(violations, rules.SinglePayloadRule.
			WithMessage("Workflow/activity methods should accept a single parameter object (use a POJO for multiple values)").
			At(rng))
	}

	primitiveCount := 0
	for _, p := range paramNodes {
		if isPrimitiveParam(p, content) {
			primitiveCount++
		}
	}
	if primitiveCount >= 2 {
		violations = append(violations, rules.PrimitiveParamsRule.
			WithMessage("Use a POJO instead of multiple primitive parameters for workflow/activity inputs").
			At(rng))
	}

	return violations
}

func isPrimitiveParam(param *sitter.Node, content []byte) bool {
	for i := 0; i < int(param.ChildCount()); i++ {
		child := param.Child(i)
		switch child.Type() {
		case rules.NodeTypeIdentifier:
			return primitiveTypes[child.Content(content)]
		case rules.NodeIntegralType, rules.NodeFloatingPoint, rules.NodeBooleanType:
			return true
		}
	}
	return false
}

func findMethodName(method *sitter.Node) *sitter.Node {
	for i := 0; i < int(method.ChildCount()); i++ {
		child := method.Child(i)
		if child.Type() == rules.NodeIdentifier {
			return child
		}
	}
	return nil
}
