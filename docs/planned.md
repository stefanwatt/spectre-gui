# Planned features

## Lua code hygiene
Inline Lua snippets scattered across Go files (e.g. `ExecLua` strings in `neovim/file_explorer.go`, `neovim/neovim.go`, etc.) should be extracted into proper `.lua` files. This makes them:
- Syntax-highlighted and lintable
- Easier to read, edit, and maintain
- Testable independently from the Go code

**Approach:** Create a `lua/` directory for app-owned Lua modules. Load them via `ExecLua(loadfile(...))` or by adding the directory to neovim's `runtimepath`. Each logical concern (autocmd bridge, mini.files helpers, markdown support, etc.) gets its own file.

## Pickers
### live grep
since include/exclude are filenames/directories, it shouldnt be just free text.
would be nice to get some sort of completion
