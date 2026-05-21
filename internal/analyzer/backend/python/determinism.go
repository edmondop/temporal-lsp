package python

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var workflowDefnDecorators = []string{
	rules.PyDecoratorWorkflowDefn,
	rules.PyDecoratorDefn,
	rules.PyDecoratorFullDefn,
}

var workflowRunDecorators = []string{
	rules.PyDecoratorWorkflowRun,
	rules.PyDecoratorRun,
	rules.PyDecoratorFullRun,
}

type DeterminismAnalyzer struct{}

func (a *DeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".py") {
		return false
	}
	s := string(content)
	return strings.Contains(s, rules.PythonSDKImportFrom) || strings.Contains(s, rules.PythonSDKImport)
}

func (a *DeterminismAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	scopes := findWorkflowRunScopes(tree.RootNode(), content)
	if len(scopes) == 0 {
		return nil, nil
	}

	var violations []rules.Violation
	for _, scope := range scopes {
		violations = append(violations, findBannedCallsInScope(scope, content)...)
	}
	return violations, nil
}

func findWorkflowRunScopes(root *sitter.Node, content []byte) []*sitter.Node {
	var scopes []*sitter.Node

	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() != rules.NodeDecoratedDef {
			continue
		}
		if !hasAnyDecorator(child, content, workflowDefnDecorators) {
			continue
		}
		classDef := rules.FindChildOfType(child, rules.NodeClassDef)
		if classDef == nil {
			continue
		}
		body := rules.FindChildOfType(classDef, rules.NodeBlock)
		if body == nil {
			continue
		}

		for j := 0; j < int(body.ChildCount()); j++ {
			method := body.Child(j)
			if method.Type() != rules.NodeDecoratedDef {
				continue
			}
			if !hasAnyDecorator(method, content, workflowRunDecorators) {
				continue
			}
			funcDef := rules.FindChildOfType(method, rules.NodeFunctionDef)
			if funcDef == nil {
				continue
			}
			funcBody := rules.FindChildOfType(funcDef, rules.NodeBlock)
			if funcBody != nil {
				scopes = append(scopes, funcBody)
			}
		}
	}
	return scopes
}

func findBannedCallsInScope(scope *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation
	rules.WalkNode(scope, func(n *sitter.Node) {
		if n.Type() != rules.NodeCall {
			return
		}
		callText := callFunctionText(n, content)
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

func callFunctionText(callNode *sitter.Node, content []byte) string {
	if callNode.ChildCount() == 0 {
		return ""
	}
	return callNode.Child(0).Content(content)
}

func hasAnyDecorator(decoratedDef *sitter.Node, content []byte, names []string) bool {
	for i := 0; i < int(decoratedDef.ChildCount()); i++ {
		child := decoratedDef.Child(i)
		if child.Type() != rules.NodeDecorator {
			continue
		}
		text := child.Content(content)
		text = strings.TrimPrefix(text, "@")
		text = strings.TrimSpace(text)
		if idx := strings.Index(text, "("); idx != -1 {
			text = text[:idx]
		}
		for _, name := range names {
			if text == name {
				return true
			}
		}
	}
	return false
}
