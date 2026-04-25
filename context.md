# Code Context

## Files Retrieved
1. `features/file-explorer/file-explorer.go` (lines 37-125, 147-217, 245-312, 341-488, 574-639) - core file-explorer state, nav flow, buffer writes, attach point, and splice logic.
2. `features/file-explorer/event-handlers.go` (lines 46-84) - `nvim_buf_lines_event` gate, current-buffer filter, detach logging.
3. `neovim/nvim_adapter.go` (lines 20-31, 54-87) - autocmd creation, `AttachBuffer`, `OpenSplitRight`, buffer/window ops.
4. `features/file-explorer/registry.go` (lines 9-105) - unused pane/buffer registry scaffold, already keyed by `bufNr` for line maps.
5. `core/projection/file_explorer_projector.go` (lines 21-55, 123-174) - UI payload contract for `parent/current/preview`.
6. `frontend/src/lib/file-explorer/FileExplorer.svelte` (lines 115-200) - frontend already renders `parent/current/preview`; schema can stay same.

## Key Code
- `FileExplorer` still model one live `current` pane, not buffer cache:
  ```go
  type FileExplorer struct {
      parent  Directory
      current Directory
      preview Directory
  }
  ```
  `Open()` creates 3 bufs once, loads dirs, then attaches only `e.current.BufNr`.

- Navigation rewrites current state in place:
  ```go
  // GoIn
  e.parent.Path = e.current.Path
  e.parent.Entries = e.current.Entries
  if err := e.loadDirectory(&e.current, newCurrentPath); err != nil { ... }
  e.clearPreview()
  if err := e.refreshVisiblePanes(); err != nil { ... }
  ```
  `GoOut()` does same swap/reload pattern.

- `nvim_buf_lines_event` only trusts current buffer:
  ```go
  buf := utils.ReflectToInt(data[0])
  if buf != e.current.BufNr { return }
  firstline := utils.ReflectToInt(data[2])
  lastline := utils.ReflectToInt(data[3])
  if err := e.UpdateEntries(firstline, lastline, linedata); err != nil { ... }
  ```

- Panic site is slice math in `UpdateEntries()`:
  ```go
  e.current.Entries = append(e.current.Entries[:firstline], e.current.Entries[lastline:]...)
  head := append(orig[:firstline:firstline], newEntries...)
  e.current.Entries = append(head, orig[lastline:]...)
  ```
  No clamp on `firstline/lastline`. If event range exceeds current slice len, panic.

- Autocmd setup is single-current-buffer by design:
  ```go
  vim.api.nvim_create_augroup("nvim-gui-file-explorer-cursor-moved", { clear = true })
  ```
  New bind clears old binds. Fine for one active buffer. Bad for cache unless rebound per switch.

- Buffer-local keymaps are only installed once in `Open()`:
  `setupKeymaps()` binds `q`, `<Right>`, `<Left>` only for `e.current.BufNr`.

- `Registry` already has buffer-keyed state:
  `directoryLineIsDir map[int]map[int]bool`, `AssignCurrentPane`, `UpdateCurrentPaneBufnr`. But no caller uses it yet.

## Architecture
- Current flow:
  `Open()` -> create 3 scratch bufs -> load parent/current dirs -> open splits -> refresh buffers -> attach current buffer -> set current window.
- Current impl assumes `current` is ephemeral and mutable. Directory IDs come from `mapDirectoryEntries()` every load, so revisit/rebind means new IDs, not persistent buffer identity.
- One-buffer-per-directory is feasible, but current code needs buffer cache + active-buffer switch.
- Minimal structural shift:
  - add cache fields: `dirsByPath` or `dirsByBuf`, plus `currentBufNr`/`currentPath`
  - stop overwriting `e.current.Entries` for every nav; switch `current` ref to cached directory/buffer for path
  - install keymaps/autocmd when buffer first created or when it becomes current
  - attach only active/current buffer; detach old active before programmatic `SetBufferLines`
  - route `nvim_buf_lines_event` by `bufNr` to matching cached directory, not hard-coded `e.current.BufNr`
  - clamp or replace splice logic; safest is full-buffer snapshot rebuild for current buffer, since buffers small
- Frontend can stay same. `FileExplorer.svelte` already renders `parent/current/preview` from payload and keys entries by `entry.id`.

## Start Here
`features/file-explorer/file-explorer.go` first. Owns state shape, nav flow, buffer writes, and current-buffer assumptions. Next: `neovim/nvim_adapter.go` for attach/detach/autocmd strategy.