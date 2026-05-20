package analyzer

type Rust struct{}

func (Rust) Analyzers() []Analyzer {
	return []Analyzer{
		NewRustDeterminismAnalyzer(),
		NewRustSignatureAnalyzer(),
		NewRustPatternAnalyzer(),
	}
}
