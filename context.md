# Code Context

## Files Retrieved
1. `features/file-explorer/file-explorer.go` (lines 36-55, 92-137, 174-218, 220-299, 323-396, 486-535, 697-767, 886-918) - core file-explorer state, handler registration, cursor math, dirty marking, pane refresh, autocmd bridge, buffer attach, buffer line writes.
2. `features/file-explorer/event-handlers.go` (lines 9-102) - direct entry points from RPC/autocmd into cursor/selection and buffer-line update flow.
3. `features/file-explorer/sync.go` (lines 43-176, 248-307) - dirty-baseline capture, sync apply, refresh, selection restore after disk reload.
4. `core/projection/file_explorer_projector.go` (lines 25-75) - `Dirty` gate, `file-explorer-update` payload, per-pane `dirty` flag source.
5. `app/runtime/runtime.go` (lines 38-93) - event loop, reducer apply, `EventFlush` projection boundary.
6. `neovim/neovim.go` (lines 80-125) - handler registration timing; file-explorer handlers registered before UI attach.
7. `transport/neovim/redraw_mapper.go` (lines 12-37) - redraw batch mapping; `flush` event preserved as batch boundary.
8. `frontend/src/lib/file-explorer/FileExplorer.svelte` (lines 83-230) - frontend render of `selectedEntryId`, `cursorCol`, dirty border, preview pane.
9. `frontend/src/app.d.ts` (lines 243-270) - frontend contract for `FileExplorerDirectory` / `FileExplorerState`.
10. `frontend/src/lib/runtime-events-service.ts` (lines 89-112) - frontend listener wiring for `file-explorer-update` / close / prompt events.
11. `features/file-explorer/cursor_test.go` (lines 74-239) - cursor/selection semantics locked by tests: visible col math, correction, row-change preserve, sync mapping.
12. `features/picker/file-watcher/file-watcher.go` (lines 30-79), `features/picker/fs.go` (lines 14-38), `frontend/src/lib/utils.service.ts` (lines 1-10), `features/markdown.go` (lines 424-462) - existing debounce/timer patterns in repo.

## Key Code
- Cursor entry flow:
  - `setupCurrentBufferAutocmd()` in `features/file-explorer/file-explorer.go` (lines 697-708) emits `FileExplorerCursorMoved` from `vim.api.nvim_win_get_cursor(0)`.
  - `onCursorMoved()` in `features/file-explorer/event-handlers.go` (lines 13-43) parses row/col, checks `GetActive()`, then calls `UpdateSelectedyEntry(row-1, col)`.
  - `UpdateSelectedyEntry()` in `features/file-explorer/file-explorer.go` (lines 182-217) updates `e.current.SelectedEntryId`, `e.current.CursorCol`, sets `e.Dirty = true`, and may correct raw cursor via `SetWindowCursor()`.

- Buffer edit flow:
  - `onBufferLines()` (lines 46-76) routes `nvim_buf_lines_event` by `bufNr` into `UpdateEntries()`.
  - `UpdateEntries()` (lines 220-279) does first-change `captureDirtyBaseline()`, splices entries, and sets `e.Dirty = true` on any text change.
  - `captureDirtyBaseline()` (lines 43-57 in `sync.go`) freezes `OriginalEntries` once per dirty buffer. Debounce here is risky: miss first snapshot, sync diff breaks.

- Refresh/sync flow:
  - `refreshVisiblePanes()` (lines 486-505) swaps buffers into windows, syncs pane buffers, then restores cursors.
  - `syncPaneCursorToSelection()` (lines 515-535) maps stored visible `CursorCol` back to raw buffer col and calls `SetWindowCursor()`.
  - `Sync()` (lines 143-176) clears `dirtyByBuf`, detaches/re-attaches tracked buffers, refreshes visible panes, then sets `e.Dirty = true` so projector pushes clean state next flush.

- Projection timing:
  - `app/runtime/runtime.go` (lines 78-92): reducer applies every event, but projector runs only when event name == `flush`.
  - `transport/neovim/redraw_mapper.go` (lines 12-37): redraw batch keeps `flush` as end-of-batch marker.
  - `core/projection/file_explorer_projector.go` (lines 25-63): if active and `e.Dirty`, emit `file-explorer-update` payload, then set `e.Dirty = false`.
  - Payload includes `dirty: p.fileExplorer.IsBufDirty(directory.BufNr)` from `dirtyByBuf` map presence, not from `e.Dirty`.

- Frontend render contract:
  - `frontend/src/app.d.ts` (lines 243-270) requires `dirty`, `selectedEntryId`, `cursorCol` on `FileExplorerDirectory`.
  - `FileExplorer.svelte` (lines 110-180) uses `selectedEntryId` to highlight row, `cursorCol` to split text at cursor, and `data.dirty` to paint yellow border/title.
  - `runtime-events-service.ts` (lines 89-112) only updates UI on `file-explorer-update` and close/prompt events; frontend is dumb, no selection math.

- Existing debounce patterns:
  - `features/picker/file-watcher/file-watcher.go` and `features/picker/fs.go` use `github.com/bep/debounce` with 100ms windows.
  - `frontend/src/lib/utils.service.ts` has `createDebounce()` via `clearTimeout` + `setTimeout`.
  - `features/markdown.go` comments show Lua timer debounce with generation guard: stop/close old timer, then `if debounce_timers[bufnr] ~= timer then return end` before callback.

## Architecture
Cursor/selection flow is split into 2 layers:

1. **Neovim event ingress**
   - Current buffer autocmd fires `FileExplorerCursorMoved` on cursor motion.
   - Neovim buffer attachment fires `nvim_buf_lines_event` on text edits.
   - Both handlers call `FileExplorer` methods directly, not runtime.

2. **Backend state + projection**
   - `UpdateSelectedyEntry()` owns semantic selection and visible cursor column.
   - `UpdateEntries()` owns live buffer mirror + dirty baseline capture.
   - Both mark `FileExplorer.Dirty = true`.
   - Runtime only emits UI after `flush`, so multiple redraw events in same batch already coalesce once.
   - `FileExplorerProjector` reads `FileExplorer` state, emits `file-explorer-update`, then clears `Dirty`.
   - Frontend receives payload and just renders rows/cursor/dirty border.

3. **Sync path**
   - `dirtyByBuf` is persistent truth for unsynced dirs.
   - `Sync()` clears `dirtyByBuf`, refreshes disk state, restores selection by path, then flips `Dirty` so UI updates cleanly.

### Safe debounce point
If debounce needed, safest place is **after semantic state update**, before UI emission, or as a small coalescer around `e.Dirty`/projection.
Do **not** debounce before `UpdateSelectedyEntry()` / `UpdateEntries()`:
- cursor correction (`SetWindowCursor`) needs immediate state
- `GoIn` / `GoOut` use current selection right away
- dirty-baseline capture needs first edit event
- `refreshVisiblePanes()` / `Sync()` can race with stale timer callbacks

### Sync risks
- `FileExplorer` has no mutex. `FileExplorerCursorMoved`, `nvim_buf_lines_event`, `GoIn`, `GoOut`, `Sync`, `Close`, and projector reads can touch same maps/fields.
- Timer/goroutine callback that mutates `FileExplorer` can race unless serialized or guarded.
- `UpdateSelectedyEntry()` may trigger `SetWindowCursor()` and re-fire cursor autocmd. Any debounce must ignore stale callbacks / reentry.
- `UpdateEntries()` uses splice ranges from event stream; batching/delaying raw line events can invalidate `firstline/lastline` unless you re-snapshot buffer content at fire time.

## Start Here
`features/file-explorer/event-handlers.go` (lines 9-102). Direct ingress from RPC/autocmd into selection + edit flow. Next open `features/file-explorer/file-explorer.go` (lines 182-217, 220-279, 697-708) for exact state mutation and cursor correction.