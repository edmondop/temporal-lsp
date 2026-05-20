package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"
)

type RustPatternAnalyzer struct{}

func NewRustPatternAnalyzer() *RustPatternAnalyzer {
	return &RustPatternAnalyzer{}
}

func (a *RustPatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".rs") {
		return false
	}
	s := string(content)
	return strings.Contains(s, "temporal_sdk") || strings.Contains(s, "temporal_client") || strings.Contains(s, "temporal_sdk_core")
}

func (a *RustPatternAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(rust.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()
	scopes := findRustWorkflowFnBodies(root, content)

	var violations []Violation
	for _, scope := range scopes {
		violations = append(violations, checkRustPatterns(scope, content)...)
	}
	return violations, nil
}

func checkRustPatterns(scope *sitter.Node, content []byte) []Violation {
	var violations []Violation
	const ref = "https://github.com/jlegrone/100-temporal-mistakes"

	walkNode(scope, func(n *sitter.Node) {
		if n.Type() != "loop_expression" {
			return
		}
		if rustHasContinueAsNew(scope, content) {
			return
		}
		violations = append(violations, Violation{
			RuleID:    "temporal/unbounded-loop",
			Message:   "Infinite loop without continue_as_new() risks history growth; add continue_as_new",
			Severity:  2,
			Range:     nodeToRange(n),
			Reference: ref,
		})
	})

	walkNode(scope, func(n *sitter.Node) {
		if n.Type() != "while_expression" {
			return
		}
		if !isRustWhileTrue(n, content) {
			return
		}
		if rustHasContinueAsNew(scope, content) {
			return
		}
		violations = append(violations, Violation{
			RuleID:    "temporal/unbounded-loop",
			Message:   "Infinite loop without continue_as_new() risks history growth; add continue_as_new",
			Severity:  2,
			Range:     nodeToRange(n),
			Reference: ref,
		})
	})

	walkNode(scope, func(n *sitter.Node) {
		if n.Type() != "call_expression" {
			return
		}
		callText := rustCallText(n, content)
		if !strings.Contains(callText, "execute_activity") {
			return
		}
		if !rustHasTimeoutInScope(scope, content) {
			violations = append(violations, Violation{
				RuleID:    "temporal/activity-timeout-required",
				Message:   "Set start_to_close_timeout or schedule_to_close_timeout in ActivityOptions",
				Severity:  2,
				Range:     nodeToRange(n),
				Reference: ref,
			})
		}
	})

	return violations
}

func isRustWhileTrue(whileNode *sitter.Node, content []byte) bool {
	for i := 0; i < int(whileNode.ChildCount()); i++ {
		child := whileNode.Child(i)
		if child.Type() == "boolean_literal" && child.Content(content) == "true" {
			return true
		}
	}
	return false
}

func rustHasContinueAsNew(scope *sitter.Node, content []byte) bool {
	found := false
	walkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == "call_expression" {
			text := rustCallText(n, content)
			if strings.Contains(text, "continue_as_new") {
				found = true
			}
		}
	})
	return found
}

func rustHasTimeoutInScope(scope *sitter.Node, content []byte) bool {
	found := false
	walkNode(scope, func(n *sitter.Node) {
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
