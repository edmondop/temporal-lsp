Output demo.gif
Set FontSize 18
Set Width 1400
Set Height 800
Set Theme "Catppuccin Mocha"
Set Shell "zsh"
Set TypingSpeed 50ms
Set Framerate 10

# Title
Type "echo 'temporal-lsp: catch anti-patterns before they hit production'"
Enter
Sleep 1.5s
Ctrl+L

# --- Python: time/sleep violations (2 errors) ---
Type "nvim internal/analyzer/testdata/python/bad_time_workflow.py"
Enter
Sleep 3s

Type ":lua vim.diagnostic.setloclist()"
Enter
Sleep 2s

Type ":lclose"
Enter
Sleep 300ms
Type ":q"
Enter
Sleep 500ms

# --- Python: Environment variable access (3 errors) ---
Type "nvim internal/analyzer/testdata/python/bad_env_workflow.py"
Enter
Sleep 3s

Type ":lua vim.diagnostic.setloclist()"
Enter
Sleep 2s

Type ":lclose"
Enter
Sleep 300ms
Type ":q"
Enter
Sleep 500ms

# --- Python: Standard logging in workflow (3 errors) ---
Type "nvim internal/analyzer/testdata/python/bad_logging_workflow.py"
Enter
Sleep 3s

Type ":lua vim.diagnostic.setloclist()"
Enter
Sleep 2s

Type ":lclose"
Enter
Sleep 300ms
Type ":q"
Enter
Sleep 500ms

# --- Clean workflow: no diagnostics ---
Type "nvim internal/analyzer/testdata/python/good_workflow.py"
Enter
Sleep 3s

Type ":lua print('Diagnostics: ' .. #vim.diagnostic.get(0))"
Enter
Sleep 1.5s
Type ":q"
Enter
Sleep 500ms

# --- Java: determinism violations ---
Type "nvim internal/analyzer/testdata/java/BadWorkflow.java"
Enter
Sleep 3s

Type ":lua vim.diagnostic.setloclist()"
Enter
Sleep 2s

Type ":lclose"
Enter
Sleep 300ms
Type ":q"
Enter
Sleep 500ms

# --- Rust: determinism violations ---
Type "nvim internal/analyzer/testdata/rust/bad_workflow.rs"
Enter
Sleep 3s

Type ":lua vim.diagnostic.setloclist()"
Enter
Sleep 2s

Type ":lclose"
Enter
Sleep 300ms
Type ":q"
Enter
Sleep 500ms

# Outro
Type "echo '4 languages supported: Python, Go, Java, Rust. PRs welcome.'"
Enter
Sleep 2s
