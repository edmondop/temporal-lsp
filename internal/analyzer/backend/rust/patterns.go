package rust

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

type PatternAnalyzer struct{}

func (a *PatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".rs") {
		return false
	}
	s := string(content)
	return strings.Contains(s, rules.RustSDKCrate) ||
		strings.Contains(s, rules.RustSDKClient) ||
		strings.Contains(s, rules.RustSDKCore)
}

func (a *PatternAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(rust.GetLanguage())

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
		if n.Type() != rules.NodeLoopExpression {
			return
		}
		if hasContinueAsNew(scope, content) {
			return
		}
		violations = append(violations, rules.Unbounded.
			WithMessage("Infinite loop without continue_as_new() risks history growth; add continue_as_new").
			At(rules.NodeToRange(n)))
	})

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeWhileExpression {
			return
		}
		if !isWhileTrue(n, content) {
			return
		}
		if hasContinueAsNew(scope, content) {
			return
		}
		violations = append(violations, rules.Unbounded.
			WithMessage("Infinite loop without continue_as_new() risks history growth; add continue_as_new").
			At(rules.NodeToRange(n)))
	})

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeCallExpression {
			return
		}
		ct := callText(n, content)
		if !strings.Contains(ct, "execute_activity") {
			return
		}
		if !hasTimeoutInScope(scope, content) {
			violations = append(violations, rules.ActivityTimeout.
				WithMessage("Set start_to_close_timeout or schedule_to_close_timeout in ActivityOptions").
				At(rules.NodeToRange(n)))
		}
	})

	return violations
}

func isWhileTrue(whileNode *sitter.Node, content []byte) bool {
	for i := 0; i < int(whileNode.ChildCount()); i++ {
		child := whileNode.Child(i)
		if child.Type() == rules.NodeBooleanLiteral && child.Content(content) == "true" {
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
		if n.Type() == rules.NodeCallExpression {
			text := callText(n, content)
			if strings.Contains(text, "continue_as_new") {
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
		if strings.Contains(text, "start_to_close_timeout") || strings.Contains(text, "schedule_to_close_timeout") {
			found = true
		}
	})
	return found
}
