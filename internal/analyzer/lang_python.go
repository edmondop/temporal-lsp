package analyzer

type Python struct{}

func (Python) Analyzers() []Analyzer {
	return []Analyzer{
		NewPythonDeterminismAnalyzer(),
		NewPythonSignatureAnalyzer(),
		NewPythonPatternAnalyzer(),
	}
}
