package dotnet

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/csharp"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var primitiveTypes = map[string]bool{
	"int":     true,
	"long":    true,
	"double":  true,
	"float":   true,
	"bool":    true,
	"string":  true,
	"byte":    true,
	"short":   true,
	"char":    true,
	"decimal": true,
	"uint":    true,
	"ulong":   true,
	"ushort":  true,
	"sbyte":   true,
}

var signatureAttributes = []string{
	rules.DotNetAttrWorkflowRun,
	rules.DotNetAttrActivity,
}

type SignatureAnalyzer struct{}

func (a *SignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".cs") {
		return false
	}
	return strings.Contains(string(content), rules.DotNetSDKImport)
}

func (a *SignatureAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(csharp.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	var violations []rules.Violation
	rules.WalkNode(tree.RootNode(), func(n *sitter.Node) {
		if n.Type() != rules.NodeMethodDeclaration {
			return
		}
		if !hasAttribute(n, content, signatureAttributes) {
			return
		}
		violations = append(violations, checkSignature(n, content)...)
	})

	return violations, nil
}

func checkSignature(method *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation

	params := rules.FindChildOfType(method, rules.NodeParameterList)
	if params == nil {
		return nil
	}

	var paramNodes []*sitter.Node
	for i := 0; i < int(params.ChildCount()); i++ {
		child := params.Child(i)
		if child.Type() == rules.NodeParameterDecl {
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
			WithMessage("Workflow/activity methods should accept a single parameter object (use a record/class for multiple values)").
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
			WithMessage("Use a record/class instead of multiple primitive parameters for workflow/activity inputs").
			At(rng))
	}

	return violations
}

func isPrimitiveParam(param *sitter.Node, content []byte) bool {
	for i := 0; i < int(param.ChildCount()); i++ {
		child := param.Child(i)
		if child.Type() == rules.NodePredefinedType {
			return primitiveTypes[child.Content(content)]
		}
		if child.Type() == rules.NodeIdentifier {
			return primitiveTypes[child.Content(content)]
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
