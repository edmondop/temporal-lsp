package typescript

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

type DeterminismAnalyzer struct{}

func (a *DeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".ts") {
		return false
	}
	return strings.Contains(string(content), rules.TypeScriptSDKImport)
}

func (a *DeterminismAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	scopes := findWorkflowFnBodies(tree.RootNode(), content)
	if len(scopes) == 0 {
		return nil, nil
	}

	var violations []rules.Violation
	for _, scope := range scopes {
		violations = append(violations, findBannedCalls(scope, content)...)
	}
	return violations, nil
}

func findWorkflowFnBodies(root *sitter.Node, content []byte) []*sitter.Node {
	var bodies []*sitter.Node
	rules.WalkNode(root, func(n *sitter.Node) {
		if n.Type() != rules.NodeExportStatement {
			return
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			switch child.Type() {
			case rules.NodeFunctionDecl:
				body := findFunctionBody(child)
				if body != nil {
					bodies = append(bodies, body)
				}
			case rules.NodeLexicalDeclaration:
				rules.WalkNode(child, func(inner *sitter.Node) {
					if inner.Type() == rules.NodeArrowFunction {
						body := findFunctionBody(inner)
						if body != nil {
							bodies = append(bodies, body)
						}
					}
				})
			}
		}
	})
	return bodies
}


func findFunctionBody(fn *sitter.Node) *sitter.Node {
	body := rules.FindChildOfType(fn, rules.NodeStatementBlock)
	if body != nil {
		return body
	}
	return rules.FindChildOfType(fn, rules.NodeBlock)
}

func findBannedCalls(scope *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation
	rules.WalkNode(scope, func(n *sitter.Node) {
		switch n.Type() {
		case rules.NodeCallExpressionTS:
		case rules.NodeNewExpression:
		case rules.NodeMemberExpression:
			if parent := n.Parent(); parent != nil {
				if parent.Type() == rules.NodeCallExpressionTS || parent.Type() == rules.NodeMemberExpression {
					return
				}
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
