package analyzer

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"
)

var rustPrimitiveTypes = map[string]bool{
	"i8":     true,
	"i16":    true,
	"i32":    true,
	"i64":    true,
	"i128":   true,
	"u8":     true,
	"u16":    true,
	"u32":    true,
	"u64":    true,
	"u128":   true,
	"f32":    true,
	"f64":    true,
	"bool":   true,
	"String": true,
	"&str":   true,
	"usize":  true,
	"isize":  true,
}

type RustSignatureAnalyzer struct{}

func NewRustSignatureAnalyzer() *RustSignatureAnalyzer {
	return &RustSignatureAnalyzer{}
}

func (a *RustSignatureAnalyzer) Supports(uri string, content []byte) bool {
	if !strings.HasSuffix(uri, ".rs") {
		return false
	}
	s := string(content)
	return strings.Contains(s, "temporal_sdk") || strings.Contains(s, "temporal_client") || strings.Contains(s, "temporal_sdk_core")
}

func (a *RustSignatureAnalyzer) Analyze(uri string, content []byte) ([]Violation, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(rust.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	root := tree.RootNode()
	var violations []Violation

	walkNode(root, func(n *sitter.Node) {
		if n.Type() != "function_item" {
			return
		}
		allAttrs := append(rustWorkflowAttributes, "activity")
		if !hasRustAttribute(n, content, allAttrs) {
			return
		}
		violations = append(violations, checkRustSignature(n, content)...)
	})

	return violations, nil
}

func checkRustSignature(fn *sitter.Node, content []byte) []Violation {
	var violations []Violation

	params := findChildOfType(fn, "parameters")
	if params == nil {
		return nil
	}

	var businessParams []*sitter.Node
	for i := 0; i < int(params.ChildCount()); i++ {
		child := params.Child(i)
		if child.Type() != "parameter" {
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

	const ref = "https://github.com/jlegrone/100-temporal-mistakes/blob/main/src/using_more_than_one_input_response_payload/"

	nameNode := findRustFnName(fn)
	startLine := int(fn.StartPoint().Row)
	startCol := int(fn.StartPoint().Column)
	if nameNode != nil {
		startLine = int(nameNode.StartPoint().Row)
		startCol = int(nameNode.StartPoint().Column)
	}

	if len(businessParams) > 1 {
		violations = append(violations, Violation{
			RuleID:    "temporal/single-payload",
			Message:   "Workflow/activity functions should accept a single struct parameter (use a struct for multiple values)",
			Severity:  2,
			Range:     Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol},
			Reference: ref,
		})
	}

	primitiveCount := 0
	for _, p := range businessParams {
		if isRustPrimitiveParam(p, content) {
			primitiveCount++
		}
	}
	if primitiveCount >= 2 {
		violations = append(violations, Violation{
			RuleID:    "temporal/primitive-params",
			Message:   "Use a struct instead of multiple primitive parameters for workflow/activity inputs",
			Severity:  2,
			Range:     Range{StartLine: startLine, StartCol: startCol, EndLine: startLine, EndCol: startCol},
			Reference: ref,
		})
	}

	return violations
}

func isRustPrimitiveParam(param *sitter.Node, content []byte) bool {
	for i := 0; i < int(param.ChildCount()); i++ {
		child := param.Child(i)
		if child.Type() == "type_identifier" || child.Type() == "primitive_type" {
			return rustPrimitiveTypes[child.Content(content)]
		}
		if child.Type() == "reference_type" {
			text := child.Content(content)
			return rustPrimitiveTypes[text]
		}
	}
	return false
}

func findRustFnName(fn *sitter.Node) *sitter.Node {
	for i := 0; i < int(fn.ChildCount()); i++ {
		child := fn.Child(i)
		if child.Type() == "identifier" {
			return child
		}
	}
	return nil
}
