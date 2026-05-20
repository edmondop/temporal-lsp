FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /temporal-lsp ./cmd/temporal-lsp/

FROM ghcr.io/charmbracelet/vhs:latest
RUN apt-get update && apt-get install -y neovim zsh && rm -rf /var/lib/apt/lists/*

COPY --from=builder /temporal-lsp /usr/local/bin/temporal-lsp

RUN mkdir -p /root/.config/nvim
COPY <<'EOF' /root/.config/nvim/init.lua
vim.opt.swapfile = false
vim.opt.shadafile = "NONE"
vim.opt.number = true
vim.opt.signcolumn = "yes"
vim.opt.termguicolors = true

-- Make diagnostic messages more visible
vim.diagnostic.config({
  virtual_text = true,
  signs = true,
  underline = true,
  severity_sort = true,
})

vim.api.nvim_create_autocmd("FileType", {
  pattern = { "python", "go", "java", "rust" },
  callback = function()
    vim.lsp.start({
      name = "temporal-lsp",
      cmd = { "/usr/local/bin/temporal-lsp" },
      root_dir = vim.fn.getcwd(),
    })
  end,
})
EOF

WORKDIR /project
COPY . .
CMD ["vhs", "demo.tape"]
