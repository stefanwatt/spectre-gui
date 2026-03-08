# Completion Menu Setup

This project includes a native Svelte completion menu that integrates with [blink.cmp](https://github.com/Saghen/blink.cmp).

## Required Configuration

To use the native completion menu, you need to configure blink.cmp in your Neovim config to disable its built-in UI.
To preserve functionality when using terminal neovim you can reference the global `nvim_gui` variable:

```lua
require('blink.cmp').setup({
  completion = {
    menu = { enabled = not vim.g.nvim_gui },
    ghost_text = { enabled = not vim.g.nvim_gui },
    documentation = { auto_show = not vim.g.nvim_gui },
  }
})
```

## How It Works

1. Blink.cmp continues to provide completion logic (LSP, buffer, path sources, fuzzy matching, etc.)
2. The native UI bridge (in Lua) listens to blink's internal emitter events
3. Completion items are sent to the Go backend via RPC
4. The Go backend emits events to the Svelte frontend
5. A native `CompletionMenu.svelte` component renders the completion items with rich styling

## Features

### Completion Menu
- **Rich UI**: Native Svelte component with kind icons, colors, and smooth rendering
- **Smart positioning**: Automatically positions below cursor, flips above when near bottom
- **Keyboard navigation**: Standard C-n, C-p, CR keymaps continue to work
- **Source indicators**: Shows which completion source provided each item (LSP, Buffer, etc.)
- **Deprecated items**: Strikethrough styling for deprecated completions

### Documentation Window
- **Automatic display**: Shows documentation for the selected completion item
- **Markdown rendering**: Renders markdown documentation with proper formatting
- **Type signatures**: Displays function signatures and type information
- **Smart positioning**: Positions to the right of the completion menu

## Testing

To test the completion menu:

1. Build and run the application: `wails dev`
2. Open a file with LSP support (e.g., a Go file if gopls is installed)
3. Enter insert mode and start typing
4. The native completion menu should appear below the cursor
5. Use C-n/C-p to navigate, CR to accept

## Architecture

**Data flow:**
```
blink.cmp (Lua) → vim.rpcnotify → Go handlers → Runtime.EventsEmit → Svelte reactive state → CompletionMenu + DocumentationWindow components
```

**Documentation flow:**
- When a completion item is selected, blink.cmp's `select_emitter` fires
- The Lua bridge extracts `item.documentation` and `item.detail`
- Documentation is sent alongside the selection index via RPC
- DocumentationWindow renders markdown or plaintext in a side panel

**Key files:**
- `neovim/completion.go` - Go structs, parser, embedded Lua bridge
- `neovim/neovim.go` - Handler registration
- `frontend/src/lib/completion/CompletionMenu.svelte` - Completion menu UI
- `frontend/src/lib/completion/DocumentationWindow.svelte` - Documentation panel UI
- `frontend/src/lib/state.svelte.ts` - Reactive completion state
