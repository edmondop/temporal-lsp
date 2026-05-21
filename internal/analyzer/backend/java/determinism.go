package java

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var workflowAnnotations = []string{
	rules.JavaAnnotationWorkflowMethod,
	rules.JavaAnnotationSignalMethod,
	rules.JavaAnnotationQueryMethod,
}

type DeterminismAnalyzer struct{}

func (a *DeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".java") {
		return false
	}
	return strings.Contains(string(content), rules.JavaSDKImport)
}

func (a *DeterminismAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

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
		if n.Type() != rules.NodeMethodDecl {
			return
		}
		if hasAnnotation(n, content, workflowAnnotations) {
			body := rules.FindChildOfType(n, rules.NodeBlock)
			if body != nil {
				bodies = append(bodies, body)
			}
		}
	})
	return bodies
}

func hasAnnotation(node *sitter.Node, content []byte, names []string) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == rules.NodeModifiers {
			for j := 0; j < int(child.ChildCount()); j++ {
				mod := child.Child(j)
				if mod.Type() == rules.NodeMarkerAnnotation || mod.Type() == rules.NodeAnnotation {
					annText := mod.Content(content)
					for _, name := range names {
						if strings.Contains(annText, name) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func findBannedCalls(scope *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation
	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeMethodInvocation && n.Type() != rules.NodeObjectCreation {
			return
		}
		callText := callText(n, content)
		for _, banned := range bannedCalls {
			for _, match := range banned.Matches {
				if strings.HasPrefix(callText, match) {
					violations = append(violations, banned.Rule.At(rules.NodeToRange(n)))
					return
				}
			}
		}
	})
	return violations
}

func callText(node *sitter.Node, content []byte) string {
	if node.Type() == rules.NodeObjectCreation {
		text := node.Content(content)
		if idx := strings.Index(text, "("); idx != -1 {
			return text[:idx]
		}
		return text
	}
	text := node.Content(content)
	if idx := strings.Index(text, "("); idx != -1 {
		return text[:idx]
	}
	return text
}
