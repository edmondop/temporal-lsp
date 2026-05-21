package analyzer

import (
	"github.com/edmondop/temporal-lsp/internal/analyzer/backend/goanalyzer"
	"github.com/edmondop/temporal-lsp/internal/analyzer/backend/java"
	"github.com/edmondop/temporal-lsp/internal/analyzer/backend/python"
	"github.com/edmondop/temporal-lsp/internal/analyzer/backend/rust"
	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
)

func NewPythonDeterminismAnalyzer() rules.Analyzer { return &python.DeterminismAnalyzer{} }
func NewPythonPatternAnalyzer() rules.Analyzer     { return &python.PatternAnalyzer{} }
func NewPythonSignatureAnalyzer() rules.Analyzer   { return &python.SignatureAnalyzer{} }

func NewJavaDeterminismAnalyzer() rules.Analyzer { return &java.DeterminismAnalyzer{} }
func NewJavaPatternAnalyzer() rules.Analyzer     { return &java.PatternAnalyzer{} }
func NewJavaSignatureAnalyzer() rules.Analyzer   { return &java.SignatureAnalyzer{} }

func NewRustDeterminismAnalyzer() rules.Analyzer { return &rust.DeterminismAnalyzer{} }
func NewRustPatternAnalyzer() rules.Analyzer     { return &rust.PatternAnalyzer{} }
func NewRustSignatureAnalyzer() rules.Analyzer   { return &rust.SignatureAnalyzer{} }

func NewGoDeterminismAnalyzer() rules.Analyzer { return &goanalyzer.DeterminismAnalyzer{} }
func NewGoPatternAnalyzer() rules.Analyzer     { return &goanalyzer.PatternAnalyzer{} }
func NewGoSignatureAnalyzer() rules.Analyzer   { return &goanalyzer.SignatureAnalyzer{} }
