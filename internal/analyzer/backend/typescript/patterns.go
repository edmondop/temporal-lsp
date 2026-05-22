package typescript

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

type PatternAnalyzer struct{}

func (a *PatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".ts") {
		return false
	}
	return strings.Contains(string(content), rules.TypeScriptSDKImport)
}

func (a *PatternAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	scopes := findWorkflowFnBodies(tree.RootNode(), content)

	var violations []rules.Violation
	for _, scope := range scopes {
		violations = append(violations, checkPatterns(scope, content)...)
	}
	return violations, nil
}

func checkPatterns(scope *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeCallExpressionTS {
			return
		}
		ct := callText(n, content)
		if strings.Contains(ct, "executeActivity") || strings.Contains(ct, "proxyActivities") {
			return
		}
		if strings.Contains(ct, "startActivity") {
			if !hasTimeoutInScope(scope, content) {
				violations = append(violations, rules.ActivityTimeout.
					WithMessage("Set startToCloseTimeout or scheduleToCloseTimeout in activity options").
					At(rules.NodeToRange(n)))
			}
		}
	})

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeWhileStatementTS {
			return
		}
		if !isWhileTrue(n, content) {
			return
		}
		if hasContinueAsNew(scope, content) {
			return
		}
		violations = append(violations, rules.Unbounded.
			WithMessage("Infinite loop without continueAsNew() risks history growth; add continueAsNew").
			At(rules.NodeToRange(n)))
	})

	return violations
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
		if n.Type() == rules.NodeCallExpressionTS {
			text := callText(n, content)
			if strings.Contains(text, "continueAsNew") {
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
		if strings.Contains(text, "startToCloseTimeout") || strings.Contains(text, "scheduleToCloseTimeout") {
			found = true
		}
	})
	return found
}
