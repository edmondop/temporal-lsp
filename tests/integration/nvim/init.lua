-- Minimal Neovim config for temporal-lsp integration testing
-- This is loaded via nvim --headless -u init.lua

-- Disable swap files and shada for headless operation
vim.opt.swapfile = false
vim.opt.shadafile = "NONE"

-- Configure the LSP client manually (no nvim-lspconfig needed)
vim.api.nvim_create_autocmd("FileType", {
  pattern = { "python", "go" },
  callback = function()
    vim.lsp.start({
      name = "temporal-lsp",
      cmd = { "/usr/local/bin/temporal-lsp" },
      root_dir = vim.fn.getcwd(),
    })
  end,
})
