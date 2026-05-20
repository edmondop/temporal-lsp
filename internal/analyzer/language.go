package analyzer

type Language interface {
	Analyzers() []Analyzer
}

func AllAnalyzers(languages ...Language) []Analyzer {
	var all []Analyzer
	for _, lang := range languages {
		all = append(all, lang.Analyzers()...)
	}
	return all
}
