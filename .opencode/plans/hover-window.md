# LSP Hover Documentation Window - Implementation Plan (v2)

## Overview
Render LSP hover documentation with rich markdown rendering. Press K → Lua gets
hover markdown directly from LSP via `buf_request_sync` → sends to Go via
`rpcnotify` → Go emits Wails event → Svelte renders with `marked`.

No neovim floating window is ever created. No race conditions.

## Architecture: Direct LSP Request
```
User presses K
  → Lua keymap handler calls vim.lsp.buf_request_sync("textDocument/hover")
  → Extracts markdown from response
  → vim.rpcnotify(channel, "LspHover", {content, row, col})
  → Go handler emits "hover-window" Wails event
  → Frontend renders markdown in HoverWindow.svelte

User moves cursor / changes mode
  → Frontend clears hover state
```

## Changes to make

### Files to revert (undo previous approach)

#### `neovim/buffer.go`
- Remove the `is_lsp_hover` win var reading (lines 72-75)

#### `core/projection/layout_content_projector.go`
- Remove `"strings"` import
- Remove `lastHoverWins` field from struct
- Remove `lastHoverWins` init in constructor
- Remove `currentHoverWins` from `projectFloatingWindows`
- Remove entire `lsp_hover` filetype check block
- Remove hover close detection loop
- Remove `extractGridText` function
- Remove `delete(currentFloatingWins, winID)` for hover windows

#### `main.go`
- Remove `hover-window-closed` event registration (keep `hover-window`)

### Files to modify

#### 1. `features/hover.go` — Replace entirely
```go
package features

// HoverDocumentation represents hover content received from LSP
type HoverDocumentation struct {
    Content string `json:"content"` // markdown content from LSP
    Row     int    `json:"row"`     // 1-based buffer line (cursor position)
    Col     int    `json:"col"`     // 0-based cursor col
}

// HoverLua sets up a K keymap that requests hover documentation directly
// from the LSP and sends the markdown content to Go via rpcnotify.
// No neovim floating window is created.
const HoverLua = `
local channel = ...
vim.keymap.set("n", "K", function()
  local params = vim.lsp.util.make_position_params()
  local cursor = vim.api.nvim_win_get_cursor(0)
  local responses = vim.lsp.buf_request_sync(0, "textDocument/hover", params, 2000)
  if not responses then return end

  for _, resp in pairs(responses) do
    if resp.result and resp.result.contents then
      local contents = resp.result.contents
      local content = ""
      if type(contents) == "string" then
        content = contents
      elseif contents.value then
        content = contents.value
      end
      if content ~= "" then
        vim.rpcnotify(channel, "LspHover", {content, cursor[1], cursor[2]})
        return
      end
    end
  end
end)
`
```

#### 2. `neovim/neovim.go`
- Change ExecLua to pass channel ID: `NvimClient.ExecLua(features.HoverLua, nil, NvimClient.ChannelID())`
- Register `LspHover` handler (before AttachUI, with other handlers):
```go
NvimClient.RegisterHandler("LspHover", func(updates ...[]interface{}) {
    for _, data := range updates {
        if len(data) < 3 {
            continue
        }
        content, _ := data[0].(string)
        row := utils.ReflectToInt(data[1])
        col := utils.ReflectToInt(data[2])
        if content == "" {
            continue
        }
        EmitEvent("hover-window", features.HoverDocumentation{
            Content: content,
            Row:     row,
            Col:     col,
        })
    }
})
```

#### 3. `main.go`
- Remove `hover-window-closed` (no longer needed — dismissal is frontend-driven)
- Keep `hover-window`

#### 4. `frontend/src/app.d.ts` — Simplify HoverWindow
```ts
interface HoverWindow {
    content: string; // markdown from LSP
    row: number;     // 1-based buffer line
    col: number;     // 0-based cursor col
}
```

#### 5. `frontend/src/lib/runtime-events-service.ts`
- Simplify `hover-window` handler (remove floating windows filtering)
- Remove `hover-window-closed` handler
- Add dismissal: clear hover on `cursor-changed` event
```ts
Events.On('hover-window', (ev) => {
    setHoverWindow(ev.data);
});

Events.On('cursor-changed', (ev) => {
    // ... existing cursor handling ...
    setHoverWindow(null);  // dismiss hover on cursor move
});
```

#### 6. `frontend/src/lib/hover/HoverWindow.svelte` — Simplify positioning
Position based on `row` from hover data. Use `document.getElementById('row-' + row)`
to find the DOM element (same pattern as CompletionMenu).
No anchor window, no anchor direction.

#### 7. `frontend/src/routes/+page.svelte`
- Remove `getHoverWindow` import (no longer needed for floating window filtering)
- Revert `floatingWindows` to simple `$derived(getFloatingWindows())`
- Keep `<HoverWindow />` rendering

## Implementation Order
1. Revert projector changes (layout_content_projector.go)
2. Revert buffer.go changes
3. Replace features/hover.go
4. Update neovim/neovim.go (handler + ExecLua with channel)
5. Simplify main.go events
6. Simplify app.d.ts type
7. Simplify runtime-events-service.ts
8. Simplify HoverWindow.svelte
9. Simplify +page.svelte
