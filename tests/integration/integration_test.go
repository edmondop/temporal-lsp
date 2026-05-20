package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dagger.io/dagger"
)

type diagnostic struct {
	Message  string `json:"message"`
	Severity int    `json:"severity"`
	Lnum     int    `json:"lnum"`
	Col      int    `json:"col"`
	Source   string `json:"source"`
	Code     string `json:"code"`
}

type expectedDiag struct {
	Code string
	Lnum int // 0-indexed line number
}

func TestNeovimPythonBadWorkflowDiagnostics(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		t.Fatalf("dagger.Connect: %v", err)
	}
	defer client.Close()

	output := runNeovimCheck(ctx, t, client, "bad_workflow.py", "check_diagnostics.lua")

	var diags []diagnostic
	if err := json.Unmarshal([]byte(output), &diags); err != nil {
		t.Fatalf("failed to parse diagnostics JSON: %v\nraw output: %s", err, output)
	}

	// bad_workflow.py has these violations at specific lines (0-indexed):
	// Line 13: datetime.datetime.now()  → temporal/no-time-now
	// Line 14: time.sleep(1)            → temporal/no-sleep
	// Line 15: random.randint(1, 10)    → temporal/no-random
	// Line 16: requests.get(...)        → temporal/no-io
	// Line 17: threading.Thread(...)    → temporal/no-goroutine
	// Line 18: threading.Lock()         → temporal/no-mutex
	// Line 19: queue.Queue()            → temporal/no-channel
	expected := []expectedDiag{
		{Code: "temporal/no-time-now", Lnum: 13},
		{Code: "temporal/no-sleep", Lnum: 14},
		{Code: "temporal/no-random", Lnum: 15},
		{Code: "temporal/no-io", Lnum: 16},
		{Code: "temporal/no-goroutine", Lnum: 17},
		{Code: "temporal/no-mutex", Lnum: 18},
		{Code: "temporal/no-channel", Lnum: 19},
	}

	if len(diags) != len(expected) {
		t.Errorf("expected exactly %d diagnostics, got %d:", len(expected), len(diags))
		for _, d := range diags {
			t.Logf("  [line %d] %s: %s", d.Lnum, d.Code, d.Message)
		}
	}

	for _, exp := range expected {
		found := false
		for _, d := range diags {
			if d.Code == exp.Code && d.Lnum == exp.Lnum {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing diagnostic: %s at line %d", exp.Code, exp.Lnum)
		}
	}

	// Verify all diagnostics come from temporal-lsp
	for _, d := range diags {
		if d.Source != "temporal-lsp" {
			t.Errorf("unexpected source %q for diagnostic %s", d.Source, d.Code)
		}
	}
}

func TestNeovimPythonBadSignaturesDiagnostics(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		t.Fatalf("dagger.Connect: %v", err)
	}
	defer client.Close()

	output := runNeovimCheck(ctx, t, client, "bad_signatures.py", "check_diagnostics.lua")

	var diags []diagnostic
	if err := json.Unmarshal([]byte(output), &diags); err != nil {
		t.Fatalf("failed to parse diagnostics JSON: %v\nraw output: %s", err, output)
	}

	// bad_signatures.py should trigger:
	// - temporal/single-payload on the workflow.run method (line 6, "run")
	// - temporal/primitive-params on the workflow.run method (line 6, "run")
	// - temporal/single-payload on the activity (line 11, "bad_activity")
	// - temporal/primitive-params on the activity (line 11, "bad_activity")
	expectedRules := map[string]int{
		"temporal/single-payload":  2,
		"temporal/primitive-params": 2,
	}

	ruleCount := map[string]int{}
	for _, d := range diags {
		ruleCount[d.Code]++
	}

	for rule, wantCount := range expectedRules {
		if ruleCount[rule] != wantCount {
			t.Errorf("expected %d %s violations, got %d", wantCount, rule, ruleCount[rule])
		}
	}
}

func TestNeovimPythonBadPatternsDiagnostics(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		t.Fatalf("dagger.Connect: %v", err)
	}
	defer client.Close()

	output := runNeovimCheck(ctx, t, client, "bad_patterns.py", "check_diagnostics.lua")

	var diags []diagnostic
	if err := json.Unmarshal([]byte(output), &diags); err != nil {
		t.Fatalf("failed to parse diagnostics JSON: %v\nraw output: %s", err, output)
	}

	expectedRules := map[string]int{
		"temporal/activity-timeout-required": 1,
		"temporal/unbounded-loop":            1,
	}

	ruleCount := map[string]int{}
	for _, d := range diags {
		ruleCount[d.Code]++
	}

	for rule, wantCount := range expectedRules {
		if ruleCount[rule] != wantCount {
			t.Errorf("expected %d %s violations, got %d", wantCount, rule, ruleCount[rule])
		}
	}
}

func TestNeovimPythonGoodWorkflowNoDiagnostics(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		t.Fatalf("dagger.Connect: %v", err)
	}
	defer client.Close()

	output := runNeovimCheck(ctx, t, client, "good_workflow.py", "check_no_diagnostics.lua")
	if output != "[]" {
		t.Errorf("expected no diagnostics for good_workflow.py, got: %s", output)
	}
}

func TestNeovimPythonGoodSignaturesNoDiagnostics(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		t.Fatalf("dagger.Connect: %v", err)
	}
	defer client.Close()

	output := runNeovimCheck(ctx, t, client, "good_signatures.py", "check_no_diagnostics.lua")
	if output != "[]" {
		t.Errorf("expected no diagnostics for good_signatures.py, got: %s", output)
	}
}

func TestNeovimPythonGoodPatternsNoDiagnostics(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		t.Fatalf("dagger.Connect: %v", err)
	}
	defer client.Close()

	output := runNeovimCheck(ctx, t, client, "good_patterns.py", "check_no_diagnostics.lua")
	if output != "[]" {
		t.Errorf("expected no diagnostics for good_patterns.py, got: %s", output)
	}
}

// Fixture validation: prove test code is real, compilable code — not gibberish.

func TestGoFixturesAreRealTemporalWorkflows(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		t.Fatalf("dagger.Connect: %v", err)
	}
	defer client.Close()

	src := client.Host().Directory(projectRoot())

	// Build the verification binary (static to avoid glibc issues)
	verifyBinary := client.Container().
		From("golang:1.26-bookworm").
		WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build")).
		WithDirectory("/src", src).
		WithWorkdir("/src/tests/integration/workers/verify_go_fixtures").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", "/out/verify", "."}).
		File("/out/verify")

	// Run Temporal + verify in a single container to avoid service binding issues
	result := client.Container().
		From("temporalio/admin-tools:1.30.4").
		WithFile("/usr/local/bin/verify", verifyBinary).
		WithExec([]string{"sh", "-c", "temporal server start-dev --ip 127.0.0.1 &\n/usr/local/bin/verify 127.0.0.1:7233"})

	output, err := result.Stdout(ctx)
	if err != nil {
		stderr, _ := result.Stderr(ctx)
		t.Fatalf("Go fixture verification failed: %v\nstderr: %s", err, stderr)
	}

	stderr, _ := result.Stderr(ctx)
	if stderr != "" {
		t.Logf("verify stderr:\n%s", stderr)
	}

	if !strings.Contains(output, "All Go fixtures verified") {
		t.Errorf("fixture verification did not fully pass:\n%s", output)
	}
	t.Log(output)
}

func TestPythonFixturesAreRealTemporalWorkflows(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		t.Fatalf("dagger.Connect: %v", err)
	}
	defer client.Close()

	src := client.Host().Directory(projectRoot())

	// Python verify uses WorkflowEnvironment.start_local() (embedded test server)
	// No external Temporal server or binary needed
	result := client.Container().
		From("python:3.12-slim").
		WithMountedCache("/root/.cache/pip", client.CacheVolume("pip")).
		WithExec([]string{"pip", "install", "temporalio", "requests"}).
		WithDirectory("/testdata", src.Directory("internal/analyzer/testdata/python")).
		WithFile("/verify.py", src.File("tests/integration/workers/verify_python_fixtures.py")).
		WithExec([]string{"python", "/verify.py"})

	output, err := result.Stdout(ctx)
	if err != nil {
		stderr, _ := result.Stderr(ctx)
		t.Fatalf("Python fixture verification failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(output, "All fixtures verified") {
		t.Errorf("fixture verification did not fully pass:\n%s", output)
	}
	t.Log(output)
}

func runNeovimCheck(ctx context.Context, t *testing.T, client *dagger.Client, testFile, luaScript string) string {
	t.Helper()

	src := client.Host().Directory(projectRoot())

	// Build the temporal-lsp binary in a Go container
	lspBinary := client.Container().
		From("golang:1.26-bookworm").
		WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build")).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"go", "build", "-o", "/out/temporal-lsp", "./cmd/temporal-lsp/"}).
		File("/out/temporal-lsp")

	// Runtime: Debian Trixie has Neovim 0.10.4 in apt (required for vim.lsp.start())
	runtime := client.Container().
		From("debian:trixie-slim").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "neovim"}).
		WithFile("/usr/local/bin/temporal-lsp", lspBinary).
		WithDirectory("/nvim-config", src.Directory("tests/integration/nvim")).
		WithDirectory("/testdata", src.Directory("internal/analyzer/testdata/python")).
		WithWorkdir("/testdata").
		WithExec([]string{
			"nvim", "--headless",
			"-u", "/nvim-config/init.lua",
			"-l", "/nvim-config/" + luaScript,
			testFile,
		})

	output, err := runtime.Stdout(ctx)
	if err != nil {
		// Try to get stderr for debugging
		stderr, _ := runtime.Stderr(ctx)
		t.Fatalf("neovim execution failed for %s with %s: %v\nstderr: %s", testFile, luaScript, err, stderr)
	}

	return strings.TrimSpace(output)
}

func projectRoot() string {
	dir, _ := os.Getwd()
	for {
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			// Check this is the main project go.mod (has cmd/ directory)
			if _, err := os.Stat(filepath.Join(dir, "cmd")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// Fallback: two levels up from tests/integration
	dir, _ = os.Getwd()
	return filepath.Dir(filepath.Dir(dir))
}
