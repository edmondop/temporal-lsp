package analyzer

type Range struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

type Fix struct {
	NewText string
}

type Violation struct {
	RuleID    string
	Message   string
	Severity  int // 1=Error, 2=Warning
	Range     Range
	Reference string
	Fix       *Fix
}

type Analyzer interface {
	Supports(uri string, content []byte) bool
	Analyze(uri string, content []byte) ([]Violation, error)
}
