package python

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

type PatternAnalyzer struct{}

func (a *PatternAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".py") {
		return false
	}
	s := string(content)
	return strings.Contains(s, rules.PythonSDKImportFrom) || strings.Contains(s, rules.PythonSDKImport)
}

func (a *PatternAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	scopes := findWorkflowRunScopes(tree.RootNode(), content)

	var violations []rules.Violation
	for _, scope := range scopes {
		violations = append(violations, checkPatterns(scope, content)...)
	}
	return violations, nil
}

func checkPatterns(scope *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeCall {
			return
		}
		callText := callFunctionText(n, content)
		if callText == "workflow.execute_activity" && !hasTimeoutArg(n, content) {
			violations = append(violations, rules.ActivityTimeout.
				WithMessage("Set start_to_close_timeout or schedule_to_close_timeout when calling execute_activity").
				At(rules.NodeToRange(n)))
		}
	})

	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeWhileStatement {
			return
		}
		if !isWhileTrue(n, content) {
			return
		}
		if hasContinueAsNewInScope(scope, content) {
			return
		}
		violations = append(violations, rules.Unbounded.
			WithMessage("Infinite loop without workflow.continue_as_new() risks history growth; add continue_as_new").
			At(rules.NodeToRange(n)))
	})

	return violations
}

func hasTimeoutArg(callNode *sitter.Node, content []byte) bool {
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
		if child.Type() == "keyword_argument" && child.ChildCount() > 0 {
			key := child.Child(0).Content(content)
			if key == "start_to_close_timeout" || key == "schedule_to_close_timeout" {
				if child.ChildCount() >= 3 {
					val := child.Child(2).Content(content)
					if val != "None" {
						return true
					}
				}
			}
		}
	}
	return false
}

func isWhileTrue(whileNode *sitter.Node, content []byte) bool {
	for i := 0; i < int(whileNode.ChildCount()); i++ {
		child := whileNode.Child(i)
		if child.Type() == rules.NodeTrue || child.Type() == rules.NodePythonTrue {
			return true
		}
		if child.Type() == rules.NodeIdentifier && child.Content(content) == rules.NodePythonTrue {
			return true
		}
	}
	return false
}

func hasContinueAsNewInScope(scope *sitter.Node, content []byte) bool {
	found := false
	rules.WalkNode(scope, func(n *sitter.Node) {
		if found {
			return
		}
		if n.Type() == rules.NodeCall {
			if callFunctionText(n, content) == "workflow.continue_as_new" {
				found = true
			}
		}
	})
	return found
}
