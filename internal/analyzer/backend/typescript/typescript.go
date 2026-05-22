package typescript

import "github.com/edmondop/temporal-lsp/internal/analyzer/rules"

func Analyzers() []rules.Analyzer {
	return []rules.Analyzer{
		&DeterminismAnalyzer{},
		&PatternAnalyzer{},
		&SignatureAnalyzer{},
	}
}
