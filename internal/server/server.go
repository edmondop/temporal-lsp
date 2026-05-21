package server

import (
	"log"
	"os"

	"github.com/edmondop/temporal-lsp/internal/analyzer/rules"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspserver "github.com/tliron/glsp/server"
)

var logger = log.New(os.Stderr, "[temporal-lsp] ", log.LstdFlags)

const serverName = "temporal-lsp"

type lspServer struct {
	analyzers []rules.Analyzer
}

func NewHandler(analyzers ...rules.Analyzer) *protocol.Handler {
	s := &lspServer{analyzers: analyzers}

	handler := &protocol.Handler{}
	handler.Initialize = s.initialize
	handler.Initialized = s.initialized
	handler.Shutdown = s.shutdown
	handler.TextDocumentDidOpen = s.textDocumentDidOpen
	handler.TextDocumentDidSave = s.textDocumentDidSave

	return handler
}

func NewServer(handler *protocol.Handler) *glspserver.Server {
	return glspserver.NewServer(handler, serverName, false)
}

func (s *lspServer) initialize(_ *glsp.Context, _ *protocol.InitializeParams) (any, error) {
	syncKind := protocol.TextDocumentSyncKindFull
	return protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: boolPtr(true),
				Change:    &syncKind,
				Save: &protocol.SaveOptions{
					IncludeText: boolPtr(true),
				},
			},
		},
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name: serverName,
		},
	}, nil
}

func (s *lspServer) initialized(_ *glsp.Context, _ *protocol.InitializedParams) error {
	return nil
}

func (s *lspServer) shutdown(_ *glsp.Context) error {
	return nil
}

func (s *lspServer) textDocumentDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	content := []byte(params.TextDocument.Text)
	s.analyze(ctx, uri, content)
	return nil
}

func (s *lspServer) textDocumentDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	if params.Text != nil {
		s.analyze(ctx, params.TextDocument.URI, []byte(*params.Text))
	}
	return nil
}

func (s *lspServer) analyze(ctx *glsp.Context, uri string, content []byte) {
	var allViolations []rules.Violation

	for _, a := range s.analyzers {
		if !a.Supports(uri, content) {
			continue
		}
		violations, err := a.Analyze(uri, content)
		if err != nil {
			logger.Printf("analyzer error for %s: %v", uri, err)
			continue
		}
		allViolations = append(allViolations, violations...)
	}

	diagnostics := make([]protocol.Diagnostic, 0, len(allViolations))
	for _, v := range allViolations {
		severity := protocol.DiagnosticSeverity(v.Severity)
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      protocol.UInteger(v.Range.StartLine),
					Character: protocol.UInteger(v.Range.StartCol),
				},
				End: protocol.Position{
					Line:      protocol.UInteger(v.Range.EndLine),
					Character: protocol.UInteger(v.Range.EndCol),
				},
			},
			Severity: &severity,
			Source:   strPtr(serverName),
			Message:  v.Message,
			Code:     &protocol.IntegerOrString{Value: string(v.RuleID)},
		})
	}

	logger.Printf("publishing %d diagnostics for %s", len(diagnostics), uri)

	ctx.Notify(string(protocol.ServerTextDocumentPublishDiagnostics), protocol.PublishDiagnosticsParams{
		URI:         protocol.DocumentUri(uri),
		Diagnostics: diagnostics,
	})
}

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}
