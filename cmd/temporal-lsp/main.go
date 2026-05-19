package main

import (
	"log"

	"github.com/edmondop/temporal-lsp/internal/analyzer"
	"github.com/edmondop/temporal-lsp/internal/server"
)

func main() {
	handler := server.NewHandler(
		analyzer.NewGoDeterminismAnalyzer(),
		analyzer.NewGoSignatureAnalyzer(),
		analyzer.NewGoPatternAnalyzer(),
		analyzer.NewPythonDeterminismAnalyzer(),
		analyzer.NewPythonSignatureAnalyzer(),
		analyzer.NewPythonPatternAnalyzer(),
	)
	srv := server.NewServer(handler)

	if err := srv.RunStdio(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
