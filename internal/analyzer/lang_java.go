package analyzer

type Java struct{}

func (Java) Analyzers() []Analyzer {
	return []Analyzer{
		NewJavaDeterminismAnalyzer(),
		NewJavaSignatureAnalyzer(),
		NewJavaPatternAnalyzer(),
	}
}
