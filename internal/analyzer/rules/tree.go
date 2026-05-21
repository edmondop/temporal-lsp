package rules

import sitter "github.com/smacker/go-tree-sitter"

func WalkNode(node *sitter.Node, fn func(*sitter.Node)) {
	fn(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		WalkNode(node.Child(i), fn)
	}
}

func FindChildOfType(node *sitter.Node, typeName string) *sitter.Node {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == typeName {
			return child
		}
	}
	return nil
}

func NodeToRange(n *sitter.Node) Range {
	return Range{
		StartLine: int(n.StartPoint().Row),
		StartCol:  int(n.StartPoint().Column),
		EndLine:   int(n.EndPoint().Row),
		EndCol:    int(n.EndPoint().Column),
	}
}
