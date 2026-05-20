package main

import (
	"log"

	"github.com/edmondop/temporal-lsp/internal/analyzer"
	"github.com/edmondop/temporal-lsp/internal/server"
)

func main() {
	analyzers := analyzer.AllAnalyzers(
		analyzer.Go{},
		analyzer.Python{},
		analyzer.Java{},
		analyzer.Rust{},
	)

	handler := server.NewHandler(analyzers...)
	srv := server.NewServer(handler)

	if err := srv.RunStdio(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
