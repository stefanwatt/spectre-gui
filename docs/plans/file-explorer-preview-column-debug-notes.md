# File Explorer Preview Column Debug Notes

## Context

Plan implemented from `docs/plans/file-explorer-preview-column-plan.md`, but result is partially broken.

User-reported current behavior:

1. Text preview sometimes stuck on `Loading preview…`.
2. After stuck text preview, navigating to directory shows directory preview.
3. Navigating back to file can keep showing old directory preview instead of file preview.
4. Navigation regression:
   - open file explorer in `~/Projects/nvim-gui`
   - `GoOut` works
   - `GoIn` makes current pane empty
   - quit then asks confirmation even though user did not intentionally edit
5. Image preview never appears.

This file explains what changed and likely failure points for fresh agent to fix incrementally.

## What Was Implemented

### Frontend type model

File: `frontend/src/app.d.ts`

Changed file explorer preview shape from loose union:

```ts
preview?: FileExplorerDirectory | FileExplorerPreview;
```

to explicit discriminated union:

```ts
type FileExplorerPreview =
  | { kind: 'directory'; title: string; directory: FileExplorerDirectory }
  | { kind: 'textFile'; title: string; content: NvimRow[] }
  | { kind: 'localImage'; title: string; path: string };
```

Also added `path` to `FileExplorerEntry` frontend type.

### Frontend component split

Added:

- `frontend/src/lib/file-explorer/DirectoryPane.svelte`
- `frontend/src/lib/file-explorer/DirectoryPreview.svelte`
- `frontend/src/lib/file-explorer/FilePreviewGrid.svelte`

Changed:

- `frontend/src/lib/file-explorer/FileExplorer.svelte`

`FileExplorer.svelte` now renders preview by `state.preview.kind`:

- `directory` -> `DirectoryPreview`
- `textFile` -> `FilePreviewGrid`
- `localImage` -> `LocalImage`

Resize observer now observes `FilePreviewGrid` content area below header, not whole preview column.

### Local image path handling

Changed:

- `frontend/src/lib/windows/markdown/LocalImage.svelte`
- `app.go`
- `fileserver.go`
- `frontend/bindings/nvim-gui/app.ts`
- stale comment in `features/markdown.go`

Old behavior:

- frontend expected `/local-image/<base64>` URLs
- backend decoded base64 path in `ReadLocalImage`
- HTTP handler served `/local-image/<base64>`

New behavior:

- frontend detects local path by heuristic:
  - `/`
  - `~`
  - `file://`
  - Windows drive path
  - UNC path
- local path passed raw to `ReadLocalImage(path)`
- backend normalizes `file://` and `~`
- backend checks image extension allowlist and returns data URL
- `/local-image/<base64>` handler removed, leaving only shared `imageExtensions` map in `fileserver.go`

### Backend preview state

Changed:

- `features/file-explorer/file-explorer.go`

Added:

```go
type PreviewKind string

const (
  PreviewKindNone PreviewKind = ""
  PreviewKindDirectory PreviewKind = "directory"
  PreviewKindTextFile PreviewKind = "textFile"
  PreviewKindLocalImage PreviewKind = "localImage"
)
```

Added state fields:

- `previewKind`
- `previewPath`
- `previewTitle`
- `previewFiletype`
- `previewTextBufNr`
- `previewGeneration`
- `previewTimer`
- `previewRows`

Added getters used by projector.

### Debounced preview refresh

Changed:

- `UpdateSelectedyEntry`

On selection change, it calls:

```go
schedulePreviewRefresh(entry)
```

Timer delay: 75ms.

Timer has generation/path-ish guard via `previewGeneration`.

Important risk: timer callback mutates `FileExplorer` from goroutine, while other handlers/projector also mutate/read it. No mutex was added. This can cause stale state or races.

### Directory preview

Function:

```go
refreshDirectoryPreview(entry DirectoryEntry)
```

Behavior:

- calls `getOrCreateDirectory(entry.Path)`
- copies returned directory into `e.preview`
- sets `previewKind = directory`
- sets preview window buffer to directory buffer
- restores current pane focus
- marks dirty
- requests redraw

### Text preview

Function:

```go
refreshTextPreview(entry DirectoryEntry)
```

Behavior:

- reads disk file up to `previewRows`
- checks first 1KB for NUL as binary detector
- huge file disables filetype highlighting
- creates/reuses scratch preview buffer only if same path and kind still text
- loads lines into scratch buffer
- detects filetype via Lua:

```lua
vim.filetype.match({ filename = filename })
```

- sets scratch buffer into preview window
- sets preview window cursor to top
- restores current pane focus
- sets `previewKind = textFile`
- projector renders preview window grid into `content`

### Image preview

Function:

```go
refreshImagePreview(entry DirectoryEntry)
```

Behavior:

- if file extension in allowlist, sets `previewKind = localImage`
- frontend should render `LocalImage path={entry.Path}`
- no nvim grid buffer needed

Supported extensions:

```go
.png .jpg .jpeg .gif .svg .webp .bmp .ico
```

### Projector

Changed:

- `core/projection/file_explorer_projector.go`

Projector now emits:

```go
"preview": p.buildPreviewPayload(state)
```

For text preview, it looks up preview window grid:

```go
winID := p.fileExplorer.GetPreviewWinID()
win := state.Editor.Screen.Windows[winID]
grid := state.Editor.Screen.Grids[win.GridID]
content := p.renderContentPreview(grid, previewFiletype)
```

If no window/grid found or grid height is zero, it emits empty content, causing frontend `Loading preview…`.

### Nvim port changes

Changed:

- `core/ports/ports.go`
- `neovim/nvim_adapter.go`

Added methods:

- `SetBufferOption`
- `DeleteBuffer`
- `SetWindowSize`
- `ExecLua`

Used by preview scratch buffer and resize.

### Preview resize

Changed:

- `app.go`
- `FileExplorer.svelte`
- `features/file-explorer/file-explorer.go`

Frontend calls:

```ts
main.App.OnFileExplorerPreviewResize(width, height)
```

Backend calls:

```go
fileExplorer.ResizePreviewPixels(width, height)
```

Conversion is hardcoded:

- cell width: 12px
- cell height: 28px

Text preview reloads after resize so row limit matches current preview height.

### Tests added/updated

Added:

- `features/file-explorer/preview_test.go`

Tests:

- read preview line limit
- binary detection
- image extension detection

Updated fake nvim in `features/file-explorer/sync_test.go` for new port methods.

## Validation Done

Passing:

```sh
go test ./features/file-explorer
nix-shell --run 'go test -vet=off ./neovim ./features/file-explorer ./core/projection'
npm run build
```

Known unrelated existing failures:

```sh
npm run check
```

fails because script expects missing `./jsconfig.json`.

```sh
nix-shell --run 'go test ./...'
```

fails from existing unrelated issues:

- `neovim/neovim.go:126` vet warning on `log.Infof`
- picker highlighting snapshots differ on whitespace spans

## Likely Causes Of New Bugs

### Bug 1: `Loading preview…` stuck for text files

Frontend shows `Loading preview…` when `state.preview.kind === 'textFile'` but `content.length === 0`.

Backend can emit empty content when projector cannot find rendered grid:

```go
winID := p.fileExplorer.GetPreviewWinID()
win, exists := state.Editor.Screen.Windows[winID]
grid := state.Editor.Screen.Grids[win.GridID]
```

Likely causes:

1. Preview refresh sets buffer and calls `redraw`, but projector runs before Neovim sends grid lines for preview window.
2. `FileExplorer.Dirty` is set false after emitting empty content; later grid redraw may not re-emit file explorer preview because `FileExplorer.Dirty` is not tied to preview window grid dirty events.
3. Preview window may not be marked/known as file explorer window in reducer/projector path, so grid changes do not trigger file explorer projection.
4. `redraw` command from timer goroutine may not guarantee UI flush ordering.
5. Timer goroutine race can set stale state without synchronized projection.

Potential fix direction:

- Keep previous preview until text grid content exists.
- Projector should not replace text preview content with empty `[]` unless preview path changed and loading state explicit.
- Mark `FileExplorer.Dirty` when preview window grid receives redraw, not only when preview state changes.
- Or avoid relying on grid redraw for first text preview and emit plain tokens immediately as fallback.
- Serialize preview timer callback through runtime/event loop or mutex.

### Bug 2: Directory preview remains when navigating back to file

Observed: stuck text -> navigate directory -> directory preview shows -> navigate back to file -> directory preview remains.

Likely causes:

1. `refreshTextPreview` returns error before setting `previewKind = textFile`, leaving old `previewKind = directory` and old `e.preview` intact.
2. Text refresh may create/set scratch buffer but projector emits empty content; frontend old directory could stay if no new `file-explorer-update` emitted.
3. Debounced timer generation race: old directory refresh can run after newer text selection, or text refresh can be skipped by generation guard.
4. `UpdateSelectedyEntry` only schedules preview on selection ID change. If current pane gets rebuilt with reused selected ID/path weirdness, preview may not update.
5. `e.previewPath` reused both as current preview path and scratch-buffer reuse guard. During transition from directory -> text, `ensurePreviewTextBuffer` uses `e.previewPath` from previous directory, so it should recreate; but errors before final assignment leave stale directory.

Potential fix direction:

- In `refreshPreviewForEntry`, set a pending preview state before branch or clear old kind on branch start.
- On text branch start, set `previewKind = textFile`, `previewPath = entry.Path`, maybe `content` loading state, then attempt nvim setup.
- On error, show text placeholder preview instead of leaving old directory.
- Add logging around every preview branch start/end/error with generation/path/kind.

### Bug 3: Navigation regression: GoOut then GoIn yields empty current pane

Repro:

1. open file explorer in `~/Projects/nvim-gui`
2. `GoOut` works
3. `GoIn` -> current pane empty
4. quit asks confirmation despite no intended edit

Likely cause introduced by directory preview using `getOrCreateDirectory(entry.Path)`.

Directory preview now reuses real editable directory objects and buffers:

```go
directory, err := e.getOrCreateDirectory(entry.Path)
e.preview = *directory
SetBufferToWindow(previewWinID, directory.BufNr)
```

This means previewing a directory creates/uses same directory buffer that may later become current pane. Possible side effects:

1. Preview directory buffer may receive edits/autocmd/buffer events and mark dirty.
2. Preview window may display buffer without being read-only; even though frontend is read-only, nvim buffer is same editable buffer.
3. `syncPaneBuffers` skips preview buffer unless kind not text/image, but directory preview can still write entries to buffer.
4. `directoriesByBuf` maps previewed directory buffer as normal directory. Later `nvim_buf_lines_event` from preview/current can mutate it unexpectedly.
5. `GoIn` does:

```go
newCurrent, err := e.getOrCreateDirectory(selectedEntry.Path)
```

If selected directory was already previewed, it reuses same directory object/buffer. If that object/buffer was made dirty/emptied by preview lifecycle or buffer events, current pane can be empty.

6. Quit confirmation likely appears because dirty tracking sees preview-created/reused directory buffer as changed (`dirtyByBuf`).

Potential fix direction:

- Directory preview must not reuse editable canonical directory buffer.
- Create separate read-only preview directory payload from filesystem entries only, no nvim buffer, no attachment, no keymaps, no dirty tracking.
- Or deep-copy directory entries into `e.preview` without registering/attaching buffer or using same BufNr.
- Do not call `getOrCreateDirectory` for directory preview; use `mapDirectoryEntries` into a dedicated `Directory{Entries: ...}` with `BufNr: 0` or a dedicated preview-only buffer not in `directoriesByBuf`.
- Preview directory must never be in `dirtyByBuf` or `directoriesByBuf`.

This is likely highest-priority fix because it breaks navigation/data model.

### Bug 4: Confirmation prompt after no intentional edit

Likely tied to Bug 3.

Dirty prompt probably checks `dirtyByBuf`. Directory preview likely causes attached buffer events or sync mismatch for previewed directory buffer, leading dirty baseline capture.

Potential fix direction:

- Ensure preview-only directory buffers never attach to `nvim_buf_lines_event` dirty path.
- Better: no nvim buffer for directory preview at all.
- Add tests for previewing directory must not change `dirtyByBuf`.

### Bug 5: Image preview never appears

Likely causes:

1. No selected image file in current directory? Need confirm by adding image fixture/manual path.
2. Image branch only checks extension; if selected image path has uppercase extension, `isPreviewImage` lowercases, so OK.
3. If selection change timer gets stale/skipped or text branch errors stale, image branch may not run.
4. Frontend `LocalImage` prop is `altText`, not `alt`; current code uses `altText`, OK.
5. Backend `ReadLocalImage` allows extension map from `fileserver.go`, OK.
6. Local path heuristic checks absolute `/`, `~`, `file://`, Windows drive, UNC. `entry.Path` from backend should be absolute because `mapDirectoryEntries` uses joined absolute directory path. OK.
7. Wails generated binding may be stale in other generated files, but active import uses `frontend/bindings/nvim-gui/app.ts`, already changed manually. If Wails uses ID-based binding, signature still string->string, so OK.
8. Main cause may be preview state not switching to `localImage` because stale directory/text timer problem or selection not changing.
9. Another possible issue: image files get `IsDir=false`, yes.

Potential fix direction:

- Add logs in `refreshPreviewForEntry`: path, ext, isDir, branch.
- Verify `file-explorer-update` payload contains `{ kind: 'localImage', path }`.
- Test frontend with forced mock state for local image preview.
- Test `ReadLocalImage('/abs/path/image.png')` returns data URL.

## Important Design Mistake In Current Implementation

Directory preview should have been pure data/read-only, but implementation reused full directory model/buffer via `getOrCreateDirectory`.

This violates plan rule:

> Directory preview: read-only list. Preview pane never focused.

It also violates implicit separation between editable current/parent panes and preview-only pane.

Fresh agent should fix this first.

## Suggested Fix Order

### Step 1: Make directory preview pure data

Change `refreshDirectoryPreview`:

- do not call `getOrCreateDirectory`
- do not call `SetBufferToWindow` for directory preview
- call `mapDirectoryEntries(entry.Path)` directly
- set `e.preview = Directory{Entries: entries, Path: entry.Path, SelectedEntryId: 0, CursorCol: 0}`
- set `previewKind = directory`
- do not attach buffer
- do not use `directoriesByBuf`
- do not mark dirty buffers

Then retest GoOut/GoIn/quit prompt.

### Step 2: Make preview refresh state transitions robust

In `refreshPreviewForEntry`:

- log generation/path/kind/branch
- clear stale preview kind at start or set pending kind
- if text/image fails, emit placeholder of same selected kind, not old kind
- ensure errors never leave old directory preview visible for selected file

### Step 3: Fix text preview loading lifecycle

Options:

- Preserve old content until new grid has content, but distinguish selected path in title/loading.
- Emit explicit loading state? Plan did not include this, but current frontend invented `Loading preview…` based on empty content.
- Mark file explorer dirty when preview window grid redraws.
- Avoid empty text preview replacing content in projector.

### Step 4: Serialize preview timer callback

Current timer mutates `FileExplorer` from goroutine. Risky.

Fix options:

- Add mutex around `FileExplorer` state reads/writes and projector getters.
- Better: route preview refresh into existing event/runtime loop.
- At minimum, guard timer callback and all preview fields with mutex.

### Step 5: Verify image preview independently

- Add known image file under test/manual directory.
- Select it and inspect emitted payload.
- Test `ReadLocalImage` directly.
- Then test `LocalImage` rendering.

## Files Changed In Implementation

Backend:

- `app.go`
- `fileserver.go`
- `core/ports/ports.go`
- `core/projection/file_explorer_projector.go`
- `features/file-explorer/file-explorer.go`
- `features/file-explorer/sync_test.go`
- `features/file-explorer/preview_test.go`
- `features/markdown.go`
- `neovim/nvim_adapter.go`

Frontend:

- `frontend/src/app.d.ts`
- `frontend/src/lib/file-explorer/FileExplorer.svelte`
- `frontend/src/lib/file-explorer/DirectoryPane.svelte`
- `frontend/src/lib/file-explorer/DirectoryPreview.svelte`
- `frontend/src/lib/file-explorer/FilePreviewGrid.svelte`
- `frontend/src/lib/windows/markdown/LocalImage.svelte`
- `frontend/bindings/nvim-gui/app.ts`

## Notes For Fresh Agent

Do not start by rewriting everything.

First fix directory preview purity. That likely fixes navigation and false dirty prompt.

Then instrument preview branch transitions. The stuck `Loading preview…` and stale directory preview likely become easier to isolate once directory preview stops mutating shared buffers.

Be careful with `FileExplorer.Dirty` and projector timing. Preview grid content arrives through Neovim redraw events after backend preview state changes; current implementation can emit text preview before grid rows exist, then never emit again.
