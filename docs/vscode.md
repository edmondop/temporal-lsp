# VS Code

Install the extension from a local `.vsix` file:

```bash
code --install-extension integrations/vscode/temporal-lsp-0.1.0.vsix
```

Then set the LSP binary path in your settings (optional if `temporal-lsp` is on your `$PATH`):

```json
{
  "temporalLsp.path": "temporal-lsp"
}
```

Diagnostics appear as squiggly underlines with full messages in the Problems panel:

![VS Code demo](../demos/vscode.gif)
