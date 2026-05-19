# Neovim

Using `nvim-lspconfig` with a custom server definition:

```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')

if not configs.temporal_lsp then
  configs.temporal_lsp = {
    default_config = {
      cmd = { 'temporal-lsp' },
      filetypes = { 'go', 'python' },
      root_dir = lspconfig.util.root_pattern('go.mod', 'pyproject.toml', '.git'),
      settings = {},
    },
  }
end

lspconfig.temporal_lsp.setup({})
```

Diagnostics appear as inline virtual text:

![Neovim demo](../demos/neovim.gif)
