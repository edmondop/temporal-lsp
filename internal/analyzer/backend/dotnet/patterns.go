package dotnet

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/csharp"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

type PatternAnalyzer struct{}

func (a *PatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".cs") {
		return false
	}
	return strings.Contains(string(content), rules.DotNetSDKImport)
}

func (a *PatternAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(csharp.GetLanguage())

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

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeInvocationExpression {
			return
		}
		ct := callText(n, content)
		if strings.Contains(ct, "ExecuteActivityAsync") || strings.Contains(ct, "ExecuteActivity") {
			if !hasTimeoutInScope(scope, content) {
				violations = append(violations, rules.ActivityTimeout.
					WithMessage("Set StartToCloseTimeout or ScheduleToCloseTimeout in ActivityOptions").
					At(rules.NodeToRange(n)))
			}
		}
	})

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeWhileStatementCS {
			return
		}
		if !isWhileTrue(n, content) {
			return
		}
		if hasContinueAsNew(scope, content) {
			return
		}
		violations = append(violations, rules.Unbounded.
			WithMessage("Infinite loop without ContinueAsNewAsync() risks history growth; add ContinueAsNew").
			At(rules.NodeToRange(n)))
	})

	return violations
}

func isWhileTrue(whileNode *sitter.Node, content []byte) bool {
	for i := 0; i < int(whileNode.ChildCount()); i++ {
		child := whileNode.Child(i)
		if child.Type() == rules.NodeBooleanLiteral && child.Content(content) == rules.NodeTrue {
			return true
		}
	}
	return false
}

func hasContinueAsNew(scope *sitter.Node, content []byte) bool {
	found := false
	rules.WalkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == rules.NodeInvocationExpression {
			text := callText(n, content)
			if strings.Contains(text, "ContinueAsNew") {
				found = true
			}
		}
	})
	return found
}

func hasTimeoutInScope(scope *sitter.Node, content []byte) bool {
	found := false
	rules.WalkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		text := n.Content(content)
		if strings.Contains(text, "StartToCloseTimeout") || strings.Contains(text, "ScheduleToCloseTimeout") {
			found = true
		}
	})
	return found
}
