package server

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/edmondop/temporal-lsp/internal/analyzer"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type notificationCollector struct {
	mu            sync.Mutex
	notifications []collectedNotification
}

type collectedNotification struct {
	Method string
	Params json.RawMessage
}

func (nc *notificationCollector) handler() jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(func(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
		if req.Notif {
			nc.mu.Lock()
			defer nc.mu.Unlock()
			var params json.RawMessage
			if req.Params != nil {
				params = *req.Params
			}
			nc.notifications = append(nc.notifications, collectedNotification{
				Method: req.Method,
				Params: params,
			})
		}
		return nil, nil
	})
}

func (nc *notificationCollector) waitFor(method string, timeout time.Duration) *collectedNotification {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		nc.mu.Lock()
		for i := range nc.notifications {
			if nc.notifications[i].Method == method {
				n := nc.notifications[i]
				nc.mu.Unlock()
				return &n
			}
		}
		nc.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func startTestServer(t *testing.T, analyzers ...analyzer.Analyzer) (*jsonrpc2.Conn, *notificationCollector) {
	t.Helper()

	clientConn, serverConn := net.Pipe()

	handler := NewHandler(analyzers...)

	rpcHandler := jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
		glspCtx := glsp.Context{
			Method: req.Method,
			Notify: func(method string, params any) {
				conn.Notify(ctx, method, params)
			},
			Call: func(method string, params any, result any) {
				conn.Call(ctx, method, params, result)
			},
		}
		if req.Params != nil {
			glspCtx.Params = *req.Params
		}
		r, validMethod, _, err := handler.Handle(&glspCtx)
		if !validMethod {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound}
		}
		if err != nil {
			return nil, err
		}
		return r, nil
	})

	sConn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(serverConn, jsonrpc2.VSCodeObjectCodec{}),
		rpcHandler,
	)

	collector := &notificationCollector{}
	cConn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(clientConn, jsonrpc2.VSCodeObjectCodec{}),
		collector.handler(),
	)

	t.Cleanup(func() {
		cConn.Close()
		sConn.Close()
	})

	return cConn, collector
}

func TestInitializeReturnsTextDocumentSyncCapability(t *testing.T) {
	client, _ := startTestServer(t)

	params := protocol.InitializeParams{}
	var rawResult json.RawMessage
	err := client.Call(context.Background(), "initialize", params, &rawResult)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	var result protocol.InitializeResult
	if err := json.Unmarshal(rawResult, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if result.Capabilities.TextDocumentSync == nil {
		t.Fatal("expected TextDocumentSync capability to be set")
	}

	if result.ServerInfo == nil || result.ServerInfo.Name != "temporal-lsp" {
		t.Fatalf("expected ServerInfo.Name to be 'temporal-lsp', got %+v", result.ServerInfo)
	}
}

type fakeAnalyzer struct {
	violations []analyzer.Violation
}

func (f *fakeAnalyzer) Supports(_ string, _ []byte) bool {
	return true
}

func (f *fakeAnalyzer) Analyze(_ string, _ []byte) ([]analyzer.Violation, error) {
	return f.violations, nil
}

func TestDidOpenPublishesDiagnosticsFromAnalyzer(t *testing.T) {
	fake := &fakeAnalyzer{
		violations: []analyzer.Violation{
			{
				RuleID:   "temporal/no-time-now",
				Message:  "Use workflow.Now() instead of time.Now()",
				Severity: 2,
				Range: analyzer.Range{
					StartLine: 5,
					StartCol:  4,
					EndLine:   5,
					EndCol:    14,
				},
			},
		},
	}

	client, collector := startTestServer(t, fake)

	// Send initialize first (required by protocol)
	var rawResult json.RawMessage
	err := client.Call(context.Background(), "initialize", protocol.InitializeParams{}, &rawResult)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	// Send didOpen notification
	err = client.Notify(context.Background(), "textDocument/didOpen", protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        "file:///tmp/test/workflow.go",
			LanguageID: "go",
			Version:    1,
			Text:       "package main\n\nimport \"time\"\n\nfunc foo() { time.Now() }",
		},
	})
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Wait for publishDiagnostics notification
	notif := collector.waitFor("textDocument/publishDiagnostics", 2*time.Second)
	if notif == nil {
		t.Fatal("expected publishDiagnostics notification, got none")
	}

	var diagParams protocol.PublishDiagnosticsParams
	if err := json.Unmarshal(notif.Params, &diagParams); err != nil {
		t.Fatalf("failed to unmarshal diagnostics: %v", err)
	}

	if diagParams.URI != "file:///tmp/test/workflow.go" {
		t.Fatalf("expected URI 'file:///tmp/test/workflow.go', got %q", diagParams.URI)
	}

	if len(diagParams.Diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagParams.Diagnostics))
	}

	diag := diagParams.Diagnostics[0]
	if diag.Message != "Use workflow.Now() instead of time.Now()" {
		t.Fatalf("unexpected message: %q", diag.Message)
	}

	if diag.Range.Start.Line != 5 || diag.Range.Start.Character != 4 {
		t.Fatalf("unexpected range start: %+v", diag.Range.Start)
	}
}
