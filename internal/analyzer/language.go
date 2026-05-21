package analyzer

import (
	"github.com/edmondop/temporal-lsp/internal/analyzer/backend/goanalyzer"
	"github.com/edmondop/temporal-lsp/internal/analyzer/backend/java"
	"github.com/edmondop/temporal-lsp/internal/analyzer/backend/python"
	"github.com/edmondop/temporal-lsp/internal/analyzer/backend/rust"
	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

func AllAnalyzers() []rules.Analyzer {
	var all []rules.Analyzer
	all = append(all, goanalyzer.Analyzers()...)
	all = append(all, python.Analyzers()...)
	all = append(all, java.Analyzers()...)
	all = append(all, rust.Analyzers()...)
	return all
}
