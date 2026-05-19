package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

type PythonPatternAnalyzer struct{}

func NewPythonPatternAnalyzer() *PythonPatternAnalyzer {
	return &PythonPatternAnalyzer{}
}

func (a *PythonPatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".py") {
		return false
	}
	s := string(content)
	return strings.Contains(s, "from temporalio") || strings.Contains(s, "import temporalio")
}

func (a *PythonPatternAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()

	var violations []Violation

	// Check workflow.run scopes for patterns
	scopes := findWorkflowRunScopes(root, content)
	for _, scope := range scopes {
		violations = append(violations, checkPythonWorkflowPatterns(scope, content)...)
	}

	return violations, nil
}

func checkPythonWorkflowPatterns(scope nodeRange, content []byte) []Violation {
	var violations []Violation
	const ref = "https://github.com/jlegrone/100-temporal-mistakes"

	walkNode(scope.node, func(n *sitter.Node) {
		if n.Type() != "call" {
			return
		}

		callText := callFunctionText(n, content)

		// activity-timeout-required: execute_activity without start_to_close_timeout
		if callText == "workflow.execute_activity" {
			if !hasTimeoutArg(n, content) {
				violations = append(violations, Violation{
					RuleID:    "temporal/activity-timeout-required",
					Message:   "Set start_to_close_timeout or schedule_to_close_timeout when calling execute_activity",
					Severity:  2,
					Range:     nodeToRange(n),
					Reference: ref,
				})
			}
		}
	})

	// unbounded-loop: while True without continue_as_new
	walkNode(scope.node, func(n *sitter.Node) {
		if n.Type() != "while_statement" {
			return
		}
		if !isWhileTrue(n, content) {
			return
		}
		if hasContinueAsNewInScope(scope.node, content) {
			return
		}
		violations = append(violations, Violation{
			RuleID:    "temporal/unbounded-loop",
			Message:   "Infinite loop without workflow.continue_as_new() risks history growth; add continue_as_new",
			Severity:  2,
			Range:     nodeToRange(n),
			Reference: ref,
		})
	})

	return violations
}

func hasTimeoutArg(callNode *sitter.Node, content []byte) bool {
	// Look for keyword arguments with timeout in the name
	for i := 0; i < int(callNode.ChildCount()); i++ {
		child := callNode.Child(i)
		if child.Type() == "argument_list" {
			return hasTimeoutKeyword(child, content)
		}
	}
	return false
}

func hasTimeoutKeyword(argList *sitter.Node, content []byte) bool {
	for i := 0; i < int(argList.ChildCount()); i++ {
		child := argList.Child(i)
		if child.Type() == "keyword_argument" {
			// First child is the key
			if child.ChildCount() > 0 {
				key := child.Child(0).Content(content)
				if key == "start_to_close_timeout" || key == "schedule_to_close_timeout" {
					// Check value is not None
					if child.ChildCount() >= 3 {
						val := child.Child(2).Content(content)
						if val != "None" {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func isWhileTrue(whileNode *sitter.Node, content []byte) bool {
	// while_statement has condition as second child (after "while" keyword)
	for i := 0; i < int(whileNode.ChildCount()); i++ {
		child := whileNode.Child(i)
		if child.Type() == "true" || child.Type() == "True" {
			return true
		}
		// In tree-sitter python, True is an identifier
		if child.Type() == "identifier" && child.Content(content) == "True" {
			return true
		}
	}
	return false
}

func hasContinueAsNewInScope(scope *sitter.Node, content []byte) bool {
	found := false
	walkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == "call" {
			callText := callFunctionText(n, content)
			if callText == "workflow.continue_as_new" {
				found = true
			}
		}
	})
	return found
}

func nodeToRange(n *sitter.Node) Range {
	return Range{
		StartLine: int(n.StartPoint().Row),
		StartCol:  int(n.StartPoint().Column),
		EndLine:   int(n.EndPoint().Row),
		EndCol:    int(n.EndPoint().Column),
	}
}
