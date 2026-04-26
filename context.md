# Code Context

## Files Retrieved
1. `features/file-explorer/file-explorer.go` (lines 36-53, 76-170, 201-256, 284-290, 568-579) - FileExplorer state, existing public getters, and global dirty flag lifecycle.
2. `features/file-explorer/sync.go` (lines 16-19, 43-56, 68-73, 121-170) - per-buffer dirty registry (`dirtyByBuf`) and sync/close paths that clear it.
3. `features/file-explorer/event-handlers.go` (lines 13-76) - buffer edit events feed `UpdateEntries`, which sets dirty state.
4. `core/projection/file_explorer_projector.go` (lines 21-53) - current consumer; reads only global `Dirty`, not pane-specific dirty status.

## Key Code
- `FileExplorer` has:
  - `Dirty bool` global projection flag
  - private `dirtyByBuf map[int]*DirDraft`
  - `Directory` values already carry `BufNr`

- Public getters today:
```go
func (e *FileExplorer) GetParent() Directory
func (e *FileExplorer) GetCurrent() Directory
func (e *FileExplorer) GetPreview() Directory
func (e *FileExplorer) GetActive() bool
func (e *FileExplorer) GetCurrentBuf() int
```

- Dirty state capture:
```go
func (e *FileExplorer) captureDirtyBaseline(bufNr int, directory *Directory)
```
  Stores original entries in `dirtyByBuf[bufNr]` on first edit.

- Dirty state write paths:
  - `UpdateSelectedyEntry` sets `e.Dirty = true` on cursor/selection change.
  - `UpdateEntries` sets `e.Dirty = true` on buffer-line edits and calls `captureDirtyBaseline`.
  - `Sync`, `HandleConfirmChoice(2)`, and `resetSessionState` clear `dirtyByBuf`.
  - `Close` resets global `Dirty` false.

- Projector currently does this only:
```go
if !p.fileExplorer.Dirty { return UIProjection{Events: []EmittedEvent{}} }
payload := map[string]any{
  "parent":  p.fileExplorer.GetParent(),
  "current": p.fileExplorer.GetCurrent(),
  "preview": p.fileExplorer.GetPreview(),
}
```
  No pane dirty fields in payload.

## Architecture
- Dirty source of truth for sync is `dirtyByBuf`, keyed by buffer number.
- `parent` / `current` panes already have stable `BufNr` via `Directory.BufNr`.
- Projector can already obtain pane buffers from `GetParent()` / `GetCurrent()`.
- Missing piece: exported query API on `FileExplorer` to expose `dirtyByBuf` state without leaking map internals.

## Start Here
`features/file-explorer/file-explorer.go` - add exported dirty-query method near `GetCurrentBuf()` / `GetParent()` / `GetCurrent()`. Projector can then ask dirty state for `parent.BufNr` and `current.BufNr`.

## Recommended API Additions
Minimal, buffer-first:
```go
func (e *FileExplorer) IsBufDirty(bufNr int) bool
```
Optional convenience wrapper:
```go
func (e *FileExplorer) IsDirectoryDirty(directory *Directory) bool
```

Recommended projector usage:
```go
parent := p.fileExplorer.GetParent()
current := p.fileExplorer.GetCurrent()
parentDirty := p.fileExplorer.IsBufDirty(parent.BufNr)
currentDirty := p.fileExplorer.IsBufDirty(current.BufNr)
```
