# temporal-lsp

## Code Style

- Self-documenting code. No comments unless the logic is genuinely non-obvious.
- No AI slop: no "// Log errors to stderr for debugging", no section-header comments restating what the next line does.
- All string literals used in logic must be named constants. No inline magic strings.
- Use typed constants for rule IDs, SDK markers, tree-sitter node types.

## Architecture

```
internal/analyzer/
  rules/           -- rule definitions, constants, shared tree-sitter utilities
  backend/
    python/        -- tree-sitter Python detection
    java/          -- tree-sitter Java detection
    rust/          -- tree-sitter Rust detection
    goanalyzer/    -- workflowcheck-based Go detection
```

- Rules are defined once in `rules/`. Each rule has an `.At(Range) Violation` factory.
- Backends return violations via `rule.At(range)` — no struct construction in detection code.
- Each language backend declares which patterns match which rule.

## Build

```bash
mise run build       # Build binary
mise run test        # Unit tests
mise run integration # Docker integration tests
mise run vscode-test # Playwright VS Code demo
```
