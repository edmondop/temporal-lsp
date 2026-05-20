package analyzer

type Go struct{}

func (Go) Analyzers() []Analyzer {
	return []Analyzer{
		NewGoDeterminismAnalyzer(),
		NewGoSignatureAnalyzer(),
		NewGoPatternAnalyzer(),
	}
}
