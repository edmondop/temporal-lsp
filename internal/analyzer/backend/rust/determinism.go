package rust

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var workflowAttributes = []string{
	rules.RustAttrWorkflowRun,
	rules.RustAttrWorkflow,
}

type DeterminismAnalyzer struct{}

func (a *DeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".rs") {
		return false
	}
	s := string(content)
	return strings.Contains(s, rules.RustSDKCrate) ||
		strings.Contains(s, rules.RustSDKClient) ||
		strings.Contains(s, rules.RustSDKCore)
}

func (a *DeterminismAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(rust.GetLanguage())

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
		if n.Type() != rules.NodeFunctionItem {
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
	parent := node.Parent()
	if parent == nil {
		return false
	}

	nodeIdx := -1
	for i := 0; i < int(parent.ChildCount()); i++ {
		if parent.Child(i) == node {
			nodeIdx = i
			break
		}
	}

	for i := nodeIdx - 1; i >= 0; i-- {
		sibling := parent.Child(i)
		if sibling.Type() == rules.NodeAttributeItem {
			attrText := sibling.Content(content)
			for _, name := range names {
				if strings.Contains(attrText, name) {
					return true
				}
			}
		} else if sibling.Type() != rules.NodeLineComment && sibling.Type() != rules.NodeBlockComment {
			break
		}
	}
	return false
}

func findBannedCalls(scope *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation
	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeCallExpression && n.Type() != rules.NodeMacroInvocation {
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
	if node.Type() == rules.NodeMacroInvocation {
		return node.Content(content)
	}
	if node.ChildCount() > 0 {
		return node.Child(0).Content(content)
	}
	return ""
}
