-- Headless Neovim script that opens a file, waits for LSP diagnostics,
-- then dumps them as JSON to stdout.
--
-- Usage: nvim --headless -u init.lua -l check_diagnostics.lua <filepath>
--
-- Exits with code 0 if diagnostics were received, 1 on timeout.

local filepath = arg[1] or vim.fn.argv(0)
if filepath == "" or filepath == nil then
  io.stderr:write("usage: nvim -l check_diagnostics.lua <filepath>\n")
  vim.cmd("cquit! 1")
  return
end

-- Open the file
vim.cmd("edit " .. vim.fn.fnameescape(filepath))

-- Wait for LSP to attach and diagnostics to arrive
local max_wait_ms = 15000
local poll_ms = 100
local waited = 0

local function check()
  local diagnostics = vim.diagnostic.get(0)
  if #diagnostics > 0 then
    return diagnostics
  end
  return nil
end

-- Poll until diagnostics arrive or timeout
while waited < max_wait_ms do
  local diags = check()
  if diags then
    -- Format diagnostics as JSON
    local results = {}
    for _, d in ipairs(diags) do
      table.insert(results, {
        message = d.message,
        severity = d.severity,
        lnum = d.lnum,
        col = d.col,
        source = d.source or "",
        code = d.code or "",
      })
    end
    -- Output JSON to stdout
    io.stdout:write(vim.fn.json_encode(results) .. "\n")
    vim.cmd("qall!")
    return
  end
  vim.wait(poll_ms, function() return false end)
  waited = waited + poll_ms
end

-- Timeout: no diagnostics received
io.stderr:write("timeout: no diagnostics received after " .. max_wait_ms .. "ms\n")
vim.cmd("cquit! 1")
