# Claude Code

## Installation

```bash
# Register the marketplace
/plugin marketplace add /path/to/temporal-lsp/integrations/claude-code

# Install the plugin
/plugin install temporal-lsp@temporal-lsp
```

Restart Claude Code after installation.

## How it works

temporal-lsp runs as an LSP server alongside your existing language servers (Pyright, gopls, rust-analyzer, etc.). It uses distinct language IDs (`temporal-python`, `temporal-go`, `temporal-java`, `temporal-rust`) to avoid collisions with other LSP plugins.

Diagnostics appear automatically when Claude Code reads or edits workflow files:

https://github.com/user-attachments/assets/demos/claude-code.mp4

## Complementary tools

temporal-lsp catches violations in existing code. For proactive guidance — teaching Claude to write correct Temporal code from the start — install the official [Temporal Developer Skill](https://github.com/temporalio/skill-temporal-developer):

```bash
/plugin install temporal-developer@claude-plugins-official
```

Or via the standalone Claude Code plugin: [temporalio/claude-temporal-plugin](https://github.com/temporalio/claude-temporal-plugin).

| | temporal-lsp | temporal-developer skill |
|---|---|---|
| **What** | Static analysis (LSP diagnostics) | Knowledge/guidance (prompt context) |
| **When** | After code is written — catches violations | Before/during writing — teaches correct patterns |
| **Where** | Any LSP client (VS Code, Neovim, Claude Code) | Claude Code only |
