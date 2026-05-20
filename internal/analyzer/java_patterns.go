package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

type JavaPatternAnalyzer struct{}

func NewJavaPatternAnalyzer() *JavaPatternAnalyzer {
	return &JavaPatternAnalyzer{}
}

func (a *JavaPatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".java") {
		return false
	}
	return strings.Contains(string(content), "io.temporal")
}

func (a *JavaPatternAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()
	scopes := findJavaWorkflowMethodBodies(root, content)

	var violations []Violation
	for _, scope := range scopes {
		violations = append(violations, checkJavaPatterns(scope, content)...)
	}
	return violations, nil
}

func checkJavaPatterns(scope *sitter.Node, content []byte) []Violation {
	var violations []Violation
	const ref = "https://github.com/jlegrone/100-temporal-mistakes"

	classBody := findEnclosingClassBody(scope)

	walkNode(scope, func(n *sitter.Node) {
		if n.Type() != "method_invocation" {
			return
		}
		text := n.Content(content)
		if !strings.Contains(text, "executeActivity") && !strings.Contains(text, "execute") {
			return
		}
		callName := javaCallText(n, content)
		if strings.HasSuffix(callName, ".execute") || strings.HasSuffix(callName, ".executeActivity") {
			searchScope := scope
			if classBody != nil {
				searchScope = classBody
			}
			if !javaHasTimeoutInScope(searchScope, content) {
				violations = append(violations, Violation{
					RuleID:    "temporal/activity-timeout-required",
					Message:   "Set withStartToCloseTimeout() or withScheduleToCloseTimeout() in ActivityOptions",
					Severity:  2,
					Range:     nodeToRange(n),
					Reference: ref,
				})
			}
		}
	})

	walkNode(scope, func(n *sitter.Node) {
		if n.Type() != "while_statement" {
			return
		}
		if !isJavaWhileTrue(n, content) {
			return
		}
		if javaHasContinueAsNew(scope, content) {
			return
		}
		violations = append(violations, Violation{
			RuleID:    "temporal/unbounded-loop",
			Message:   "Infinite loop without Workflow.continueAsNew() risks history growth; add continueAsNew",
			Severity:  2,
			Range:     nodeToRange(n),
			Reference: ref,
		})
	})

	return violations
}

func findEnclosingClassBody(node *sitter.Node) *sitter.Node {
	for n := node.Parent(); n != nil; n = n.Parent() {
		if n.Type() == "class_body" {
			return n
		}
	}
	return nil
}

func javaHasTimeoutInScope(scope *sitter.Node, content []byte) bool {
	found := false
	walkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == "method_invocation" {
			text := n.Content(content)
			if strings.Contains(text, "withStartToCloseTimeout") || strings.Contains(text, "withScheduleToCloseTimeout") {
				found = true
			}
		}
	})
	return found
}

func isJavaWhileTrue(whileNode *sitter.Node, content []byte) bool {
	cond := findChildOfType(whileNode, "parenthesized_expression")
	if cond == nil {
		return false
	}
	condText := strings.TrimSpace(cond.Content(content))
	return condText == "(true)"
}

func javaHasContinueAsNew(scope *sitter.Node, content []byte) bool {
	found := false
	walkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == "method_invocation" {
			text := javaCallText(n, content)
			if strings.Contains(text, "continueAsNew") {
				found = true
			}
		}
	})
	return found
}
