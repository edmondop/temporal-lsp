package dotnet

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/csharp"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var workflowAttributes = []string{
	rules.DotNetAttrWorkflowRun,
}

type DeterminismAnalyzer struct{}

func (a *DeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".cs") {
		return false
	}
	return strings.Contains(string(content), rules.DotNetSDKImport)
}

func (a *DeterminismAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(csharp.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	scopes := findWorkflowMethodBodies(tree.RootNode(), content)
	if len(scopes) == 0 {
		return nil, nil
	}

	var violations []rules.Violation
	for _, scope := range scopes {
		violations = append(violations, findBannedCalls(scope, content)...)
	}
	return violations, nil
}

func findWorkflowMethodBodies(root *sitter.Node, content []byte) []*sitter.Node {
	var bodies []*sitter.Node
	rules.WalkNode(root, func(n *sitter.Node) {
		if n.Type() != rules.NodeMethodDeclaration {
			return
		}
		if hasAttribute(n, content, workflowAttributes) {
			body := rules.FindChildOfType(n, rules.NodeBlock)
			if body != nil {
				bodies = append(bodies, body)
			}
		}
	})
	return bodies
}

func hasAttribute(node *sitter.Node, content []byte, names []string) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == rules.NodeAttributeList {
			attrText := child.Content(content)
			for _, name := range names {
				if strings.Contains(attrText, name) {
					return true
				}
			}
		}
	}
	return false
}

func findBannedCalls(scope *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation
	rules.WalkNode(scope, func(n *sitter.Node) {
		switch n.Type() {
		case rules.NodeInvocationExpression, rules.NodeObjectCreationExpr:
		case rules.NodeMemberAccessExpression:
			if parent := n.Parent(); parent != nil && parent.Type() == rules.NodeInvocationExpression {
				return
			}
		default:
			return
		}
		ct := callText(n, content)
		for _, banned := range bannedCalls {
			for _, match := range banned.Matches {
				if strings.Contains(ct, match) {
					violations = append(violations, banned.Rule.At(rules.NodeToRange(n)))
					return
				}
			}
		}
	})
	return violations
}

func callText(node *sitter.Node, content []byte) string {
	text := node.Content(content)
	if idx := strings.Index(text, "("); idx != -1 {
		return text[:idx]
	}
	return text
}
