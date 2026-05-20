package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"
)

var rustBannedCalls = []struct {
	RuleID  string
	Message string
	Matches []string
}{
	{
		RuleID:  "temporal/no-time-now",
		Message: "Use workflow time utilities instead of SystemTime/Instant in workflows",
		Matches: []string{"SystemTime::now", "Instant::now", "Utc::now", "Local::now"},
	},
	{
		RuleID:  "temporal/no-sleep",
		Message: "Use Temporal's timer API instead of thread::sleep or tokio::time::sleep in workflows",
		Matches: []string{"thread::sleep", "tokio::time::sleep", "sleep("},
	},
	{
		RuleID:  "temporal/no-random",
		Message: "Use workflow random utilities instead of rand in workflows",
		Matches: []string{"rand::", "thread_rng", "OsRng", "StdRng"},
	},
	{
		RuleID:  "temporal/no-io",
		Message: "Move IO operations to an activity",
		Matches: []string{"std::fs::", "File::open", "File::create", "TcpStream::", "reqwest::", "std::net::"},
	},
	{
		RuleID:  "temporal/no-goroutine",
		Message: "Use Temporal's async primitives instead of spawning threads/tasks in workflows",
		Matches: []string{"thread::spawn", "tokio::spawn", "tokio::task::spawn"},
	},
	{
		RuleID:  "temporal/no-mutex",
		Message: "Temporal workflows are single-threaded; remove Mutex/RwLock usage",
		Matches: []string{"Mutex::new", "RwLock::new"},
	},
}

type RustDeterminismAnalyzer struct{}

func NewRustDeterminismAnalyzer() *RustDeterminismAnalyzer {
	return &RustDeterminismAnalyzer{}
}

func (a *RustDeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".rs") {
		return false
	}
	s := string(content)
	return strings.Contains(s, "temporal_sdk") || strings.Contains(s, "temporal_client") || strings.Contains(s, "temporal_sdk_core")
}

func (a *RustDeterminismAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(rust.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()
	scopes := findRustWorkflowFnBodies(root, content)
	if len(scopes) == 0 {
		return nil, nil
	}

	var violations []Violation
	for _, scope := range scopes {
		violations = append(violations, findRustBannedCalls(scope, content)...)
	}
	return violations, nil
}

var rustWorkflowAttributes = []string{"workflow_run", "workflow"}

func findRustWorkflowFnBodies(root *sitter.Node, content []byte) []*sitter.Node {
	var bodies []*sitter.Node
	walkNode(root, func(n *sitter.Node) {
		if n.Type() != "function_item" {
			return
		}
		if hasRustAttribute(n, content, rustWorkflowAttributes) {
			body := findChildOfType(n, "block")
			if body != nil {
				bodies = append(bodies, body)
			}
		}
	})
	return bodies
}

func hasRustAttribute(node *sitter.Node, content []byte, names []string) bool {
	parent := node.Parent()
	if parent == nil {
		return false
	}

	nodeIdx := -1
	for i := 0; i < int(parent.ChildCount()); i++ {
		if parent.Child(i) == node {
			nodeIdx = i
			break
		}
	}

	// attribute_item nodes precede the function_item as siblings
	for i := nodeIdx - 1; i >= 0; i-- {
		sibling := parent.Child(i)
		if sibling.Type() == "attribute_item" {
			attrText := sibling.Content(content)
			for _, name := range names {
				if strings.Contains(attrText, name) {
					return true
				}
			}
		} else if sibling.Type() != "line_comment" && sibling.Type() != "block_comment" {
			break
		}
	}
	return false
}

func findRustBannedCalls(scope *sitter.Node, content []byte) []Violation {
	var violations []Violation
	walkNode(scope, func(n *sitter.Node) {
		if n.Type() != "call_expression" && n.Type() != "macro_invocation" {
			return
		}
		callText := rustCallText(n, content)
		for _, banned := range rustBannedCalls {
			for _, match := range banned.Matches {
				if strings.Contains(callText, match) {
					violations = append(violations, Violation{
						RuleID:   banned.RuleID,
						Message:  banned.Message,
						Severity: 1,
						Range:    nodeToRange(n),
					})
					return
				}
			}
		}
	})
	return violations
}

func rustCallText(node *sitter.Node, content []byte) string {
	if node.Type() == "macro_invocation" {
		return node.Content(content)
	}
	// call_expression: first child is the function path
	if node.ChildCount() > 0 {
		return node.Child(0).Content(content)
	}
	return ""
}
