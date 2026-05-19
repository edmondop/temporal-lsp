-- Headless Neovim script that opens a file, waits briefly for LSP,
-- then asserts no diagnostics are reported.
--
-- Usage: nvim --headless -u init.lua -l check_no_diagnostics.lua <filepath>
--
-- Outputs "[]" and exits 0 if no diagnostics after waiting period.
-- Exits 1 if unexpected diagnostics appear.

local filepath = arg[1] or vim.fn.argv(0)
if filepath == "" or filepath == nil then
  io.stderr:write("usage: nvim -l check_no_diagnostics.lua <filepath>\n")
  vim.cmd("cquit! 1")
  return
end

-- Open the file
vim.cmd("edit " .. vim.fn.fnameescape(filepath))

-- Wait for LSP to attach and give it time to report diagnostics
local wait_ms = 5000
local poll_ms = 100
local waited = 0

while waited < wait_ms do
  vim.wait(poll_ms, function() return false end)
  waited = waited + poll_ms
end

-- Check diagnostics
local diagnostics = vim.diagnostic.get(0)
if #diagnostics == 0 then
  io.stdout:write("[]\n")
  vim.cmd("qall!")
else
  -- Unexpected diagnostics found
  local results = {}
  for _, d in ipairs(diagnostics) do
    table.insert(results, {
      message = d.message,
      severity = d.severity,
      lnum = d.lnum,
      col = d.col,
      source = d.source or "",
      code = d.code or "",
    })
  end
  io.stdout:write(vim.fn.json_encode(results) .. "\n")
  vim.cmd("cquit! 1")
end
