package java

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

type PatternAnalyzer struct{}

func (a *PatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".java") {
		return false
	}
	return strings.Contains(string(content), rules.JavaSDKImport)
}

func (a *PatternAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	scopes := findWorkflowMethodBodies(tree.RootNode(), content)

	var violations []rules.Violation
	for _, scope := range scopes {
		violations = append(violations, checkPatterns(scope, content)...)
	}
	return violations, nil
}

func checkPatterns(scope *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation

	classBody := findEnclosingClassBody(scope)

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeMethodInvocation {
			return
		}
		text := n.Content(content)
		if !strings.Contains(text, "executeActivity") && !strings.Contains(text, "execute") {
			return
		}
		callName := callText(n, content)
		if strings.HasSuffix(callName, ".execute") || strings.HasSuffix(callName, ".executeActivity") {
			searchScope := scope
			if classBody != nil {
				searchScope = classBody
			}
			if !hasTimeoutInScope(searchScope, content) {
				violations = append(violations, rules.ActivityTimeout.
					WithMessage("Set withStartToCloseTimeout() or withScheduleToCloseTimeout() in ActivityOptions").
					At(rules.NodeToRange(n)))
			}
		}
	})

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeWhileStatement {
			return
		}
		if !isWhileTrue(n, content) {
			return
		}
		if hasContinueAsNew(scope, content) {
			return
		}
		violations = append(violations, rules.Unbounded.
			WithMessage("Infinite loop without Workflow.continueAsNew() risks history growth; add continueAsNew").
			At(rules.NodeToRange(n)))
	})

	return violations
}

func findEnclosingClassBody(node *sitter.Node) *sitter.Node {
	for n := node.Parent(); n != nil; n = n.Parent() {
		if n.Type() == rules.NodeClassBody {
			return n
		}
	}
	return nil
}

func hasTimeoutInScope(scope *sitter.Node, content []byte) bool {
	found := false
	rules.WalkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == rules.NodeMethodInvocation {
			text := n.Content(content)
			if strings.Contains(text, "withStartToCloseTimeout") || strings.Contains(text, "withScheduleToCloseTimeout") {
				found = true
			}
		}
	})
	return found
}

func isWhileTrue(whileNode *sitter.Node, content []byte) bool {
	cond := rules.FindChildOfType(whileNode, rules.NodeParenthesized)
	if cond == nil {
		return false
	}
	condText := strings.TrimSpace(cond.Content(content))
	return condText == "(true)"
}

func hasContinueAsNew(scope *sitter.Node, content []byte) bool {
	found := false
	rules.WalkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == rules.NodeMethodInvocation {
			text := callText(n, content)
			if strings.Contains(text, "continueAsNew") {
				found = true
			}
		}
	})
	return found
}
