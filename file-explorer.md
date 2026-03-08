# nvim-gui file explorer

## concept
- always three columns like ranger 
- miller columns
- always full screen overlay (except status line)
- left column is one dir up
- right column is preview


## features
- select several items by pressing space. items can then by yanked, deleted etc together
- visual mode 
### preview
- right column always shows a preview
- if selected entry in center column is a file then show file preview
  file preview should be syntax hightlighted properly. this can be achieved by also creating a buffer for it and setting filetype
- if selected entry in center column is a dir then show dir contents
  this dir preview can just be pretty much the same display as the center or left column. 
### edit as buffer
center column has an underlying buffer such that files/dirs can be renamed, created, deleted
by manipulating the neovim buffer. just like oil.nvim or mini.files
any text manipulation action should work as expected. 
deleting a line deletes the file/dir.
yanking a line copies the  file/dir.
y2<Up> copies the current file/dir and two entries above.
visual mode works for selecting multiple entries to be yanked/deleted later

you can rename a file by simply changing  the text.

any number of changes can be made before then confirming with "=" 

# File Explorer Implementation Plan

## Context
Implement a ranger-style Miller columns file explorer as a full-screen overlay. The center column is backed by a neovim buffer so all text editing (rename, delete lines, add lines) works natively. The right column uses a neovim buffer for syntax-highlighted file previews. Changes are applied atomically when the user presses `=`.

## Architecture Overview

```
┌──────────────┬──────────────────────────┬──────────────────┐
│  Parent Dir  │  Current Dir (nvim buf)  │  Preview         │
│  (Svelte)    │  editable buffer         │  (nvim buf/list) │
│  ~25% width  │  ~50% width              │  ~25% width      │
└──────────────┴──────────────────────────┴──────────────────┘
```

- **Left column**: Go sends parent dir entries → Svelte renders a simple list
- **Center column**: Neovim scratch buffer in a floating window → editable, all vim motions work
- **Right column**: Neovim floating window for file preview (syntax hl) OR Svelte list for dir preview

The center and preview floating windows flow through the existing grid rendering pipeline. The frontend filters them out of normal `FloatingWindowContainer` rendering by filetype and routes their content to the `FileExplorer.svelte` overlay.

## Implementation Steps

### 1. Go: `neovim/file_explorer.go` — Core struct and state

**New file.** Contains:

```go
type FileExplorer struct {
    screen          *Screen
    Active          bool
    CurrentDir      string
    centerBuf       nvim.Buffer
    centerWin       nvim.Window
    centerGridId    int
    previewBuf      nvim.Buffer
    previewWin      nvim.Window
    previewGridId   int
    originalEntries []FileEntry   // matches buffer lines by extmark
    namespace       int           // extmark namespace
}

type FileEntry struct {
    Name  string `json:"name"`
    Path  string `json:"path"`
    IsDir bool   `json:"isDir"`
}
```

Key methods:

- **`Open(dir string)`**
  1. `os.ReadDir(dir)` → sort dirs first, then files, alphabetically
  2. `CreateBuffer(false, true)` → scratch buffer
  3. Set buffer lines to entry names (dirs get trailing `/`)
  4. Place extmarks on each line (namespace "file_explorer") to track identity
  5. `OpenWindow(buf, true, &WindowConfig{Relative: "editor", ...})` for center (~50% width, centered)
  6. Set buffer options: `buftype=nofile`, `bufhidden=wipe`
  7. Create preview buffer + floating window (right ~25%)
  8. Set buffer-local keymaps via lua (h, l, Space, =, q)
  9. Set `CursorMoved` autocmd → `vim.rpcnotify(channel, 'fe-cursor-moved', row)`
  10. Emit `file-explorer-open` event to frontend with parent entries, currentDir

- **`Close()`** — close windows, wipe buffers, emit `file-explorer-close`

- **`UpdatePreview(row int)`** — called on cursor move
  - Look up entry at row in originalEntries (via extmark)
  - If dir: emit `file-explorer-preview` with dir listing
  - If file: read file (capped at ~200 lines), set preview buffer lines + filetype → neovim renders it

- **`NavigateInto()`** — enter selected dir: change currentDir, refresh buffer
- **`NavigateUp()`** — go to parent dir: change currentDir, refresh buffer

- **`Apply()`** — diff and execute:
  1. Read current buffer lines
  2. Get all extmarks with current positions
  3. For each extmark: compare current line text with original entry name
     - Text changed → `os.Rename(oldPath, newPath)`
     - Extmark gone (line deleted) → `os.RemoveAll(path)` (with confirmation?)
  4. Lines without extmarks → new files: `os.Create(name)` or `os.MkdirAll(name)` (if ends with `/`)
  5. Refresh all columns

- **`ToggleSelect(row int)`** — toggle highlight on line (using extmark hl_group)

### 2. Go: Integration points

**`neovim/neovim.go`**:
- Add `FileExplorer *FileExplorer` field to appropriate struct
- Initialize in startup
- Register RPC handlers: `fe-cursor-moved`, `fe-navigate-into`, `fe-navigate-up`, `fe-apply`, `fe-close`, `fe-toggle-select`

**`neovim/keymaps.go`**:
- Register `<leader>e` → opens file explorer at current file's directory

**`neovim/floating-windows.go` or `screen.go`**:
- In `EmitFloatingWindows()`: no changes needed — frontend filters by filetype

**Buffer-local keymaps** (set via lua in `Open()`):
| Key | Action |
|-----|--------|
| `h` | `vim.rpcnotify(ch, 'fe-navigate-up')` |
| `l` / `<CR>` | `vim.rpcnotify(ch, 'fe-navigate-into')` |
| `<Space>` | `vim.rpcnotify(ch, 'fe-toggle-select')` |
| `=` | `vim.rpcnotify(ch, 'fe-apply')` |
| `q` / `<Esc>` | `vim.rpcnotify(ch, 'fe-close')` |

### 3. Frontend: Types (`frontend/src/app.d.ts`)

```typescript
interface FileExplorerState {
    open: boolean;
    currentDir: string;
    parentEntries: FileEntry[];
    previewEntries: FileEntry[];  // for dir preview
    previewIsFile: boolean;
    cursorRow: number;
    selectedLines: number[];
}

interface FileEntry {
    name: string;
    path: string;
    isDir: boolean;
}
```

### 4. Frontend: State (`frontend/src/lib/state.svelte.ts`)

Add:
```typescript
export let fileExplorer = $state<App.FileExplorerState>({
    open: false,
    currentDir: '',
    parentEntries: [],
    previewEntries: [],
    previewIsFile: false,
    cursorRow: 0,
    selectedLines: [],
});

export let fileExplorerCenterWindowId = $state(0);
export let fileExplorerPreviewWindowId = $state(0);
```

### 5. Frontend: Events (`frontend/src/lib/runtime-events-service.ts`)

New event listeners:
- `file-explorer-open` → set `fileExplorer.open = true`, populate state, store window IDs
- `file-explorer-close` → set `fileExplorer.open = false`
- `file-explorer-preview` → update `fileExplorer.previewEntries` or `previewIsFile`

In the `floating_windows` handler: filter windows whose filetype starts with `"file-explorer"` and store their content separately for the file explorer overlay.

### 6. Frontend: `FileExplorer.svelte` component

**New file:** `frontend/src/lib/file-explorer/FileExplorer.svelte`

- Full-screen overlay, z-index 100 (above windows, below cmdline)
- CSS grid: `grid-template-columns: 1fr 2fr 1fr`
- **Header**: shows current directory path as breadcrumbs
- **Left column**: renders `parentEntries` as a styled list (highlight dirs vs files, highlight entry matching currentDir name)
- **Center column**: renders the center floating window's tokenized content (same as FloatingGrid — iterate rows/tokens, render spans with CSS classes)
- **Right column**: if `previewIsFile` → render preview floating window content; else → render `previewEntries` as styled list
- Cursor row highlighting in center column

### 7. Frontend: Integration in `+page.svelte`

Add `FileExplorer` component conditionally:
```svelte
{#if fileExplorer.open}
    <FileExplorer />
{/if}
```

### 8. Frontend: Filter floating windows

In the floating_windows event handler or in FloatingWindowContainer, skip windows that belong to the file explorer (by comparing window IDs with `fileExplorerCenterWindowId`/`fileExplorerPreviewWindowId`).

## Key Files to Create/Modify

| File | Action |
|------|--------|
| `neovim/file_explorer.go` | **CREATE** — core logic |
| `neovim/neovim.go` | MODIFY — init FileExplorer, register handlers |
| `neovim/keymaps.go` | MODIFY — add `<leader>e` binding |
| `frontend/src/lib/file-explorer/FileExplorer.svelte` | **CREATE** — overlay component |
| `frontend/src/lib/state.svelte.ts` | MODIFY — add fileExplorer state |
| `frontend/src/lib/runtime-events-service.ts` | MODIFY — add event listeners, filter floating windows |
| `frontend/src/app.d.ts` | MODIFY — add types |
| `frontend/src/routes/+page.svelte` | MODIFY — render FileExplorer overlay |

## Verification

1. Press `<leader>e` → file explorer opens as full-screen overlay showing current file's directory
2. j/k moves cursor in center column, preview updates on right
3. h goes to parent dir, l enters selected dir or opens file
4. Edit a filename in center → press `=` → file renamed on disk
5. `dd` a line → press `=` → file deleted
6. `o` + type name → press `=` → new file created
7. `q` closes the explorer
8. Build with `nix-shell --run "go build -tags webkit2_41 ./..."` to verify Go compiles

 7 tasks (3 done, 1 in progress, 3 open) · ctrl+t to hide tasks
  ◼ Add frontend types, state, and events for file explorer
  ◻ Create FileExplorer.svelte component
  ◻ Integrate FileExplorer in +page.svelte
  ◻ Build and verify compilation
  ✔ Create neovim/file_explorer.go — core struct and logic
  ✔ Modify neovim/neovim.go — init FileExplorer, register RPC handlers
  ✔ Modify neovim/keymaps.go — add leader-e binding
