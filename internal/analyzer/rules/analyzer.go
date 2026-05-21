package rules

type Analyzer interface {
	Supports(uri string, content []byte) bool
	Analyze(uri string, content []byte) ([]Violation, error)
}
