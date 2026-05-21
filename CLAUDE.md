# temporal-lsp

An LSP server that detects Temporal workflow anti-patterns across Python, Java, Rust, and Go.

## Code Style

- Self-documenting code. No comments unless the logic is genuinely non-obvious.
- No AI slop: no "// Log errors to stderr for debugging", no section-header comments restating what the next line does.
- All string literals used in logic must be named constants. No inline magic strings.
- Use typed constants for rule IDs, SDK markers, tree-sitter node types.
- Go code follows standard `gofmt`/`go vet`. Short functions (<30 lines). Early returns over nested if/else.

## Architecture

```
cmd/temporal-lsp/       -- binary entrypoint
internal/
  analyzer/
    rules/              -- rule IDs, Rule type with .At(Range) factory, shared tree utilities
    backend/
      python/           -- tree-sitter Python detection (determinism, patterns, signatures)
      java/             -- tree-sitter Java detection
      rust/             -- tree-sitter Rust detection
      goanalyzer/       -- go/analysis + workflowcheck-based Go detection
  server/               -- LSP protocol handling (glsp)
tests/
  fixtures/e2e/         -- golden fixture tests (per-language, JSON snapshots)
  vscode/               -- Playwright integration tests for VS Code extension
integrations/
  vscode/               -- VS Code extension source
  claude-code/          -- Claude Code LSP plugin config
```

- Rules are defined once in `rules/`. Each rule has an `.At(Range) Violation` factory.
- Backends return violations via `rule.At(range)` — no struct construction in detection code.
- Each language backend exposes `Analyzers() []rules.Analyzer`.
- Adding a new rule: define in `rules/rules.go`, add detection in the backend, add e2e fixture.

## Build

```bash
mise run build       # Build binary
mise run test        # Unit tests
mise run integration # Docker integration tests
mise run vscode-test # Playwright VS Code demo
```

## Testing

- Unit tests live alongside production code in `internal/analyzer/*_test.go`.
- E2e golden tests in `tests/fixtures/e2e/` — one fixture file per rule per language.
- Run: `go test ./tests/fixtures/e2e/`
- Update goldens after intentional changes: `go test ./tests/fixtures/e2e/ -update`
- `TestCoverageAllRulesHaveFixtures` enforces every declared rule has a fixture. If you add a rule, add a fixture or CI fails.
- Go fixtures require a full module (`go.mod` + `go.sum`) since the Go analyzer uses `packages.LoadAllSyntax`.

## Adding a New Rule

1. Add the rule ID constant to `internal/analyzer/rules/rules.go`
2. Add a `Rule` var with severity and reference
3. Implement detection in the relevant backend(s) under `internal/analyzer/backend/`
4. Add the rule to `expectedRulesByLanguage` in `tests/fixtures/e2e/e2e_test.go`
5. Create a fixture file in `tests/fixtures/e2e/<lang>/` that triggers the rule
6. Run `go test ./tests/fixtures/e2e/ -update` to generate the golden JSON

## Adding a New Language

1. Create `internal/analyzer/backend/<lang>/` with determinism, patterns, signatures analyzers
2. Each analyzer implements `rules.Analyzer` (Supports + Analyze)
3. Add `<lang>.go` exposing `Analyzers() []rules.Analyzer`
4. Wire into `internal/analyzer/language.go` AllAnalyzers()
5. Add fixtures under `tests/fixtures/e2e/<lang>/`
6. Add the language to `expectedRulesByLanguage` in the e2e test
