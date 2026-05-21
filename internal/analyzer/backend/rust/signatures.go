package rust

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

var primitiveTypes = map[string]bool{
	"i8": true, "i16": true, "i32": true, "i64": true, "i128": true,
	"u8": true, "u16": true, "u32": true, "u64": true, "u128": true,
	"f32": true, "f64": true,
	"bool": true, "String": true, "&str": true,
	"usize": true, "isize": true,
}

var signatureAttributes = []string{
	rules.RustAttrWorkflowRun,
	rules.RustAttrWorkflow,
	rules.RustAttrActivity,
}

type SignatureAnalyzer struct{}

func (a *SignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".rs") {
		return false
	}
	s := string(content)
	return strings.Contains(s, rules.RustSDKCrate) ||
		strings.Contains(s, rules.RustSDKClient) ||
		strings.Contains(s, rules.RustSDKCore)
}

func (a *SignatureAnalyzer) Analyze(uri string, content []byte) ([]rules.Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(rust.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	var violations []rules.Violation
	rules.WalkNode(tree.RootNode(), func(n *sitter.Node) {
		if n.Type() != rules.NodeFunctionItem {
			return
		}
		if !hasAttribute(n, content, signatureAttributes) {
			return
		}
		violations = append(violations, checkSignature(n, content)...)
	})

	return violations, nil
}

func checkSignature(fn *sitter.Node, content []byte) []rules.Violation {
	var violations []rules.Violation

	params := rules.FindChildOfType(fn, rules.NodeParameters)
	if params == nil {
		return nil
	}

	var businessParams []*sitter.Node
	for i := 0; i < int(params.ChildCount()); i++ {
		child := params.Child(i)
		if child.Type() != rules.NodeParameter {
			continue
		}
		paramText := child.Content(content)
		if strings.Contains(paramText, "self") {
			continue
		}
		if strings.Contains(paramText, "ctx") || strings.Contains(paramText, "context") || strings.Contains(paramText, "Context") {
			continue
		}
		businessParams = append(businessParams, child)
	}

	nameNode := findFnName(fn)
	startLine := int(fn.StartPoint().Row)
	startCol := int(fn.StartPoint().Column)
	if nameNode != nil {
		startLine = int(nameNode.StartPoint().Row)
		startCol = int(nameNode.StartPoint().Column)
	}
	rng := rules.Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol}

	if len(businessParams) > 1 {
		violations = append(violations, rules.SinglePayloadRule.
			WithMessage("Workflow/activity functions should accept a single struct parameter (use a struct for multiple values)").
			At(rng))
	}

	primitiveCount := 0
	for _, p := range businessParams {
		if isPrimitiveParam(p, content) {
			primitiveCount++
		}
	}
	if primitiveCount >= 2 {
		violations = append(violations, rules.PrimitiveParamsRule.
			WithMessage("Use a struct instead of multiple primitive parameters for workflow/activity inputs").
			At(rng))
	}

	return violations
}

func isPrimitiveParam(param *sitter.Node, content []byte) bool {
	for i := 0; i < int(param.ChildCount()); i++ {
		child := param.Child(i)
		switch child.Type() {
		case rules.NodeTypeIdentifier, rules.NodePrimitiveType:
			return primitiveTypes[child.Content(content)]
		case rules.NodeReferenceType:
			return primitiveTypes[child.Content(content)]
		}
	}
	return false
}

func findFnName(fn *sitter.Node) *sitter.Node {
	for i := 0; i < int(fn.ChildCount()); i++ {
		child := fn.Child(i)
		if child.Type() == rules.NodeIdentifier {
			return child
		}
	}
	return nil
}
