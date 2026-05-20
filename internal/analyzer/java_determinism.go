package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

var javaBannedCalls = []struct {
	RuleID  string
	Message string
	Matches []string
}{
	{
		RuleID:  "temporal/no-time-now",
		Message: "Use Workflow.currentTimeMillis() instead of system time in workflows",
		Matches: []string{"System.currentTimeMillis", "System.nanoTime", "Instant.now", "LocalDateTime.now", "LocalDate.now", "ZonedDateTime.now", "new Date"},
	},
	{
		RuleID:  "temporal/no-sleep",
		Message: "Use Workflow.sleep() instead of Thread.sleep() in workflows",
		Matches: []string{"Thread.sleep"},
	},
	{
		RuleID:  "temporal/no-random",
		Message: "Use Workflow.newRandom() instead of Math.random() or Random in workflows",
		Matches: []string{"Math.random", "new Random", "ThreadLocalRandom.current"},
	},
	{
		RuleID:  "temporal/no-io",
		Message: "Move network/IO calls to an activity",
		Matches: []string{"new File", "Files.", "new FileInputStream", "new FileOutputStream", "URL.openConnection", "HttpClient.", "new Socket"},
	},
	{
		RuleID:  "temporal/no-goroutine",
		Message: "Use Async.function() or Async.procedure() instead of raw threads in workflows",
		Matches: []string{"new Thread", "Executors.", "executor.submit", "executor.execute", "CompletableFuture.supplyAsync", "CompletableFuture.runAsync"},
	},
	{
		RuleID:  "temporal/no-mutex",
		Message: "Temporal workflows are single-threaded; remove lock/synchronized usage",
		Matches: []string{"new ReentrantLock", "new Semaphore", "new CountDownLatch"},
	},
}

type JavaDeterminismAnalyzer struct{}

func NewJavaDeterminismAnalyzer() *JavaDeterminismAnalyzer {
	return &JavaDeterminismAnalyzer{}
}

func (a *JavaDeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".java") {
		return false
	}
	return strings.Contains(string(content), "io.temporal")
}

func (a *JavaDeterminismAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(java.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()
	scopes := findJavaWorkflowMethodBodies(root, content)
	if len(scopes) == 0 {
		return nil, nil
	}

	var violations []Violation
	for _, scope := range scopes {
		violations = append(violations, findJavaBannedCalls(scope, content)...)
	}
	return violations, nil
}

var javaWorkflowAnnotations = []string{"WorkflowMethod", "SignalMethod", "QueryMethod"}

func findJavaWorkflowMethodBodies(root *sitter.Node, content []byte) []*sitter.Node {
	var bodies []*sitter.Node
	walkNode(root, func(n *sitter.Node) {
		if n.Type() != "method_declaration" {
			return
		}
		if hasJavaAnnotation(n, content, javaWorkflowAnnotations) {
			body := findChildOfType(n, "block")
			if body != nil {
				bodies = append(bodies, body)
			}
		}
	})
	return bodies
}

func hasJavaAnnotation(node *sitter.Node, content []byte, names []string) bool {
	parent := node.Parent()
	if parent == nil {
		return false
	}

	// Annotations live inside "modifiers" child of method_declaration
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "modifiers" {
			for j := 0; j < int(child.ChildCount()); j++ {
				mod := child.Child(j)
				if mod.Type() == "marker_annotation" || mod.Type() == "annotation" {
					annText := mod.Content(content)
					for _, name := range names {
						if strings.Contains(annText, name) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func findJavaBannedCalls(scope *sitter.Node, content []byte) []Violation {
	var violations []Violation
	walkNode(scope, func(n *sitter.Node) {
		if n.Type() != "method_invocation" && n.Type() != "object_creation_expression" {
			return
		}
		callText := javaCallText(n, content)
		for _, banned := range javaBannedCalls {
			for _, match := range banned.Matches {
				if strings.HasPrefix(callText, match) {
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

func javaCallText(node *sitter.Node, content []byte) string {
	if node.Type() == "object_creation_expression" {
		// "new Foo(...)" — extract "new Foo"
		text := node.Content(content)
		if idx := strings.Index(text, "("); idx != -1 {
			return text[:idx]
		}
		return text
	}
	// method_invocation: object.method(args)
	// We want "object.method"
	text := node.Content(content)
	if idx := strings.Index(text, "("); idx != -1 {
		return text[:idx]
	}
	return text
}
