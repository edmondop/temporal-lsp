package typescript

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var primitiveTypes = map[string]bool{
	"string":  true,
	"number":  true,
	"boolean": true,
	"bigint":  true,
}

type SignatureAnalyzer struct{}

func (a *SignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".ts") {
		return false
	}
	return strings.Contains(string(content), rules.TypeScriptSDKImport)
}

func (a *SignatureAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	var violations []rules.Violation
	rules.WalkNode(tree.RootNode(), func(n *sitter.Node) {
		if n.Type() != rules.NodeExportStatement {
			return
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			if child.Type() == rules.NodeFunctionDecl {
				violations = append(violations, checkSignature(child, content)...)
			}
		}
	})

	return violations, nil
}

func checkSignature(fn *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation

	params := findFormalParams(fn)
	if params == nil {
		return nil
	}

	var paramNodes []*sitter.Node
	for i := 0; i < int(params.ChildCount()); i++ {
		child := params.Child(i)
		if child.Type() == rules.NodeRequiredParameter || child.Type() == rules.NodeOptionalParameter {
			paramNodes = append(paramNodes, child)
		}
	}

	nameNode := findFnName(fn)
	startLine := int(fn.StartPoint().Row)
	startCol := int(fn.StartPoint().Column)
	if nameNode != nil {
		startLine = int(nameNode.StartPoint().Row)
		startCol = int(nameNode.StartPoint().Column)
	}
	rng := rules.Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol}

	if len(paramNodes) > 1 {
		violations = append(violations, rules.SinglePayloadRule.
			WithMessage("Workflow functions should accept a single object parameter (use an interface/type for multiple values)").
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
			WithMessage("Use an interface/type instead of multiple primitive parameters for workflow inputs").
			At(rng))
	}

	return violations
}

func isPrimitiveParam(param *sitter.Node, content []byte) bool {
	for i := 0; i < int(param.ChildCount()); i++ {
		child := param.Child(i)
		if child.Type() == rules.NodeTypeAnnotation {
			typeText := strings.TrimSpace(child.Content(content))
			typeText = strings.TrimPrefix(typeText, ":")
			typeText = strings.TrimSpace(typeText)
			return primitiveTypes[typeText]
		}
	}
	return false
}

func findFormalParams(fn *sitter.Node) *sitter.Node {
	return rules.FindChildOfType(fn, rules.NodeFormalParametersTS)
}

func findFnName(fn *sitter.Node) *sitter.Node {
	for i := 0; i < int(fn.ChildCount()); i++ {
		child := fn.Child(i)
		if child.Type() == rules.NodeIdentifier {
			return child
		}
	}
	return nil
}
