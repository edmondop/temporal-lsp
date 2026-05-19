package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

var bannedCalls = []struct {
	RuleID  string
	Message string
	Matches []string // prefixes to match against call text
}{
	{
		RuleID:  "temporal/no-time-now",
		Message: "Use workflow.now() instead of datetime.now() or time.time() in workflows",
		Matches: []string{"datetime.datetime.now", "datetime.now", "time.time"},
	},
	{
		RuleID:  "temporal/no-sleep",
		Message: "Use workflow.sleep() instead of time.sleep() in workflows",
		Matches: []string{"time.sleep"},
	},
	{
		RuleID:  "temporal/no-random",
		Message: "Use workflow.random() instead of random.* in workflows",
		Matches: []string{"random."},
	},
	{
		RuleID:  "temporal/no-io",
		Message: "Move network/IO calls to an activity",
		Matches: []string{"requests.", "urllib.", "open("},
	},
	{
		RuleID:  "temporal/no-goroutine",
		Message: "Use workflow.start_activity() or asyncio tasks managed by the workflow instead of threading",
		Matches: []string{"threading.Thread", "threading.start", "multiprocessing.Process", "multiprocessing.Pool"},
	},
	{
		RuleID:  "temporal/no-mutex",
		Message: "Temporal workflows are single-threaded; remove lock usage",
		Matches: []string{"threading.Lock", "threading.RLock", "asyncio.Lock"},
	},
	{
		RuleID:  "temporal/no-channel",
		Message: "Use workflow signals or workflow queues instead of queue/multiprocessing primitives",
		Matches: []string{"queue.Queue", "multiprocessing.Queue", "asyncio.Queue"},
	},
	{
		RuleID:  "temporal/no-env-access",
		Message: "Environment variables are non-deterministic; pass configuration as workflow input",
		Matches: []string{"os.getenv", "os.environ"},
	},
	{
		RuleID:  "temporal/no-standard-logging",
		Message: "Use workflow.logger instead of standard logging (avoids duplicate messages during replay)",
		Matches: []string{"logging.", "logger.", "print"},
	},
}

type PythonDeterminismAnalyzer struct{}

func NewPythonDeterminismAnalyzer() *PythonDeterminismAnalyzer {
	return &PythonDeterminismAnalyzer{}
}

func (a *PythonDeterminismAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".py") {
		return false
	}
	s := string(content)
	return strings.Contains(s, "from temporalio") || strings.Contains(s, "import temporalio")
}

// workflowDefnDecorators are all decorator forms that mark a class as a Temporal workflow.
var workflowDefnDecorators = []string{"workflow.defn", "defn", "temporalio.workflow.defn"}

// workflowRunDecorators are all decorator forms that mark a method as the workflow entry point.
var workflowRunDecorators = []string{"workflow.run", "run", "temporalio.workflow.run"}

func (a *PythonDeterminismAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()

	// Find workflow run method scopes
	scopes := findWorkflowRunScopes(root, content)
	if len(scopes) == 0 {
		return nil, nil
	}

	// Find all call expressions within workflow scopes
	var violations []Violation
	for _, scope := range scopes {
		violations = append(violations, findBannedCallsInScope(scope, content)...)
	}

	return violations, nil
}

type nodeRange struct {
	node *sitter.Node
}

func findWorkflowRunScopes(root *sitter.Node, content []byte) []nodeRange {
	var scopes []nodeRange

	// Walk top-level looking for decorated class definitions
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() != "decorated_definition" {
			continue
		}

		// Check if any decorator is a workflow defn variant
		if !hasAnyDecorator(child, content, workflowDefnDecorators) {
			continue
		}

		// Find the class_definition inside
		classDef := findChildOfType(child, "class_definition")
		if classDef == nil {
			continue
		}

		// Find methods decorated with @workflow.run
		body := findChildOfType(classDef, "block")
		if body == nil {
			continue
		}

		for j := 0; j < int(body.ChildCount()); j++ {
			method := body.Child(j)
			if method.Type() != "decorated_definition" {
				continue
			}
			if !hasAnyDecorator(method, content, workflowRunDecorators) {
				continue
			}
			funcDef := findChildOfType(method, "function_definition")
			if funcDef == nil {
				continue
			}
			funcBody := findChildOfType(funcDef, "block")
			if funcBody != nil {
				scopes = append(scopes, nodeRange{node: funcBody})
			}
		}
	}

	return scopes
}

func findBannedCallsInScope(scope nodeRange, content []byte) []Violation {
	var violations []Violation
	walkNode(scope.node, func(n *sitter.Node) {
		if n.Type() != "call" {
			return
		}
		callText := callFunctionText(n, content)
		for _, banned := range bannedCalls {
			for _, match := range banned.Matches {
				if strings.HasPrefix(callText, match) {
					violations = append(violations, Violation{
						RuleID:   banned.RuleID,
						Message:  banned.Message,
						Severity: 1,
						Range: Range{
							StartLine: int(n.StartPoint().Row),
							StartCol:  int(n.StartPoint().Column),
							EndLine:   int(n.EndPoint().Row),
							EndCol:    int(n.EndPoint().Column),
						},
					})
					return
				}
			}
		}
	})
	return violations
}

func callFunctionText(callNode *sitter.Node, content []byte) string {
	// The function being called is the first child of a call node
	if callNode.ChildCount() == 0 {
		return ""
	}
	funcNode := callNode.Child(0)
	return funcNode.Content(content)
}

func hasDecorator(decoratedDef *sitter.Node, content []byte, name string) bool {
	for i := 0; i < int(decoratedDef.ChildCount()); i++ {
		child := decoratedDef.Child(i)
		if child.Type() == "decorator" {
			text := child.Content(content)
			// decorator text includes the @
			if strings.Contains(text, name) {
				return true
			}
		}
	}
	return false
}

func hasAnyDecorator(decoratedDef *sitter.Node, content []byte, names []string) bool {
	for i := 0; i < int(decoratedDef.ChildCount()); i++ {
		child := decoratedDef.Child(i)
		if child.Type() != "decorator" {
			continue
		}
		// Extract just the decorator name (strip @ and any arguments)
		text := child.Content(content)
		text = strings.TrimPrefix(text, "@")
		text = strings.TrimSpace(text)
		// Remove anything after ( for decorators with arguments
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

func findChildOfType(node *sitter.Node, typeName string) *sitter.Node {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == typeName {
			return child
		}
	}
	return nil
}

func walkNode(node *sitter.Node, fn func(*sitter.Node)) {
	fn(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		walkNode(node.Child(i), fn)
	}
}
