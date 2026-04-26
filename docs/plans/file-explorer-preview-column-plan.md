# File Explorer Preview Column Plan

## Goal

Make file explorer preview column useful.

Rules:

- If current selected entry is directory: show directory contents in preview column.
- If current selected entry is text file: render syntax-highlighted Neovim grid from visible preview window.
- If current selected entry is local image: render image directly in web UI.
- Focus always stays in current pane. Parent and preview never receive focus after setup/refresh.
- Preview window dimensions must match frontend preview grid content area, not whole column/header.

## Current Problem

Preview currently never renders correctly because backend/frontend payload shapes disagree.

Current backend:

- `core/projection/file_explorer_projector.go` sends `preview` as flat directory pane payload via `buildPanePayload(preview, ...)`.

Current frontend:

- `frontend/src/lib/file-explorer/FileExplorer.svelte` expects `state.preview.directory` or `state.preview.content`.

Result:

- Neither branch matches reliably.
- Preview column appears empty.

## Decisions From Design Interview

### Preview kinds

Use explicit discriminated union:

```ts
type FileExplorerPreview =
  | { kind: 'directory'; title: string; directory: FileExplorerDirectory }
  | { kind: 'textFile'; title: string; content: NvimRow[] }
  | { kind: 'localImage'; title: string; path: string };
```

### Directory preview

- Read-only list.
- Render same basic directory entry style as parent/current, but no editing/cursor behavior.
- Preview pane never focused.

### Text file preview

- Use visible Neovim preview window so normal grid redraw pipeline provides cells.
- Do **not** use hidden buffer/window.
- File starts at top.
- Always load disk snapshot into scratch preview buffer.
- Do not reuse existing modified buffer for same file.
- Follow mini.files buffer lifecycle:
  - one scratch buffer for current preview path while valid
  - delete/recreate when preview path changes
  - wipe preview buffers on close
- Read only preview-window-height lines, not whole file.
- Detect binary by checking first 1KB for NUL.
- Binary placeholder: show readable placeholder row, not raw bytes.
- Skip/limit highlighting for huge files.
- Detect filetype from selected path (`vim.filetype.match({ filename = path })` or equivalent).
- Render final content using `Grid.svelte` through thin preview wrapper.

### Image preview

- Add local image preview kind.
- Supported image extensions stay: `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.webp`, `.bmp`, `.ico`.
- Frontend renders local image directly using `LocalImage`/`ReadLocalImage`.
- No Neovim grid needed for image content.

### LocalImage cleanup

- Remove base64 mechanism.
- Use raw local path in Wails `ReadLocalImage(path)`.
- `LocalImage` determines local-vs-remote with heuristic:
  - local if starts `/`, `~`, `file://`, Windows drive path, or UNC path
  - remote URL rendered directly as `src`
- Keep extension allowlist only. No workspace sandbox for now.
- Full cleanup requested:
  - remove/rework dead `/local-image/<base64>` HTTP handler
  - remove base64 comments/stale code

### Components

Replace `FileExplorer.svelte` pane snippet with components:

- `DirectoryPane.svelte`
  - used for parent/current
  - supports current cursor display/edit visuals
  - supports dirty indicator/title pill
- `DirectoryPreview.svelte`
  - read-only directory preview list
- `FilePreviewGrid.svelte`
  - thin wrapper around `Grid.svelte`
  - owns fill/overflow/background/loading layout
- `LocalImage` reused for local image preview

### Title/header

Preview has header/title for directory/text/image previews.

Resize observer measures **grid content area below header** for text preview sizing.

### Debounce

- Selection state updates immediately.
- Current pane highlight/cursor updates immediately.
- Preview rebuild is debounced by 75ms.
- Backend Go timer owns debounce.
- Timer must use generation/path guard and active check to avoid stale preview work.
- Keep previous preview visible until new preview is ready.

## Implementation Plan

### 1. Update frontend types

File: `frontend/src/app.d.ts`

Change preview model to discriminated union.

Add path to file explorer entry if missing in frontend type:

```ts
interface FileExplorerEntry {
  id: number;
  isDir: boolean;
  text: string;
  path: string;
  icon: string;
  iconClass: string;
}
```

Define:

```ts
interface FileExplorerDirectoryPreview {
  kind: 'directory';
  title: string;
  directory: FileExplorerDirectory;
}

interface FileExplorerTextFilePreview {
  kind: 'textFile';
  title: string;
  content: NvimRow[];
}

interface FileExplorerLocalImagePreview {
  kind: 'localImage';
  title: string;
  path: string;
}

type FileExplorerPreview =
  | FileExplorerDirectoryPreview
  | FileExplorerTextFilePreview
  | FileExplorerLocalImagePreview;
```

Then:

```ts
interface FileExplorerState {
  parent?: FileExplorerDirectory;
  current?: FileExplorerDirectory;
  preview?: FileExplorerPreview;
  currentWinMode?: VimMode;
}
```

### 2. Split file explorer panes into components

Files to add:

- `frontend/src/lib/file-explorer/DirectoryPane.svelte`
- `frontend/src/lib/file-explorer/DirectoryPreview.svelte`
- `frontend/src/lib/file-explorer/FilePreviewGrid.svelte`

Update:

- `frontend/src/lib/file-explorer/FileExplorer.svelte`

Remove snippet. Use components:

```svelte
<DirectoryPane data={state.parent} role="parent" />
<DirectoryPane data={state.current} role="current" />
```

Preview branch:

```svelte
{#if state.preview?.kind === 'directory'}
  <DirectoryPreview title={state.preview.title} directory={state.preview.directory} />
{:else if state.preview?.kind === 'textFile'}
  <FilePreviewGrid title={state.preview.title} content={state.preview.content} bind:contentEl={previewGridEl} />
{:else if state.preview?.kind === 'localImage'}
  <PreviewHeader title={state.preview.title} />
  <LocalImage url={state.preview.path} alt={state.preview.title} />
{/if}
```

Exact component API can differ. Important: ResizeObserver observes text-grid content area, not header.

### 3. Rework LocalImage raw-path loading

Files:

- `frontend/src/lib/windows/markdown/LocalImage.svelte`
- `app.go`
- `fileserver.go`
- generated bindings if needed
- stale comments in `features/markdown.go`

Backend `ReadLocalImage`:

- Accept raw path.
- Normalize `file://` if present.
- Expand `~` if desired.
- Keep extension allowlist.
- Read file.
- Return data URL.

Frontend `LocalImage`:

- If local path: call `ReadLocalImage(path)`.
- Else: use URL directly.
- Remove `/local-image/` prefix/base64 logic.

Cleanup:

- Delete dead `LocalImageHandler` or remove base64 code if handler still required later.
- Move shared `imageExtensions` out of deleted handler if needed.

### 4. Add backend preview state

File: `features/file-explorer/file-explorer.go`

Add fields roughly:

```go
type PreviewKind string

const (
  PreviewKindNone PreviewKind = ""
  PreviewKindDirectory PreviewKind = "directory"
  PreviewKindTextFile PreviewKind = "textFile"
  PreviewKindLocalImage PreviewKind = "localImage"
)

type FileExplorer struct {
  // existing...
  previewKind PreviewKind
  previewPath string
  previewTitle string
  previewTextBufNr int
  previewGeneration uint64
  previewTimer *time.Timer
}
```

For directory preview, can reuse existing `e.preview Directory`, but now it must mean directory preview only.

For text preview, `previewTextBufNr` is scratch buffer currently displayed in preview window.

For image preview, no nvim buffer update needed beyond maybe placeholder/empty preview window decision. Since frontend renders image directly, Neovim preview split may still exist for layout/redraw but content not used.

### 5. Trigger debounced preview refresh on selection change

File: `features/file-explorer/file-explorer.go`

In `UpdateSelectedyEntry`:

- Update selected entry immediately.
- Mark `Dirty = true` for current pane update.
- If selected path changed, schedule preview update after 75ms.

Pseudo:

```go
func (e *FileExplorer) schedulePreviewRefresh(entry DirectoryEntry) {
  e.previewGeneration++
  generation := e.previewGeneration
  path := entry.Path

  if e.previewTimer != nil {
    e.previewTimer.Stop()
  }

  e.previewTimer = time.AfterFunc(75*time.Millisecond, func() {
    if !e.active { return }
    if generation != e.previewGeneration { return }
    e.refreshPreviewForPath(path, generation)
  })
}
```

Need avoid unsafely mutating shared state from timer if broader runtime is not thread-safe. If races appear, route timer callback into existing event loop/channel or guard `FileExplorer` with mutex. This is one implementation risk to verify before coding.

### 6. Implement preview refresh by selected entry kind

File: `features/file-explorer/file-explorer.go`

Function shape:

```go
func (e *FileExplorer) refreshPreviewForEntry(entry DirectoryEntry) error
```

Branch:

#### Directory

- `getOrCreateDirectory(entry.Path)` or load read-only separate directory view if editing current shared directory state is risky.
- Copy entries into `e.preview`.
- `e.previewKind = PreviewKindDirectory`
- `e.previewPath = entry.Path`
- `e.previewTitle = entry.Text`
- Set preview window buffer to directory preview buffer if needed.
- Do not focus preview window.
- Mark `Dirty = true`.

#### Image

- `e.previewKind = PreviewKindLocalImage`
- `e.previewPath = entry.Path`
- `e.previewTitle = entry.Text`
- Optionally clear preview Neovim window buffer or show small placeholder.
- Mark `Dirty = true`.

#### Text file

- Binary/huge checks.
- Create scratch buffer for path if path changed.
- Read up to preview window height rows.
- Set lines.
- Detect filetype.
- Set buffer `filetype`.
- Set `modifiable=false`, maybe `readonly=true`.
- Set buffer into `previewWinID` via `SetBufferToWindow`.
- Ensure current pane regains focus via `SetCurrentWindow(currentWinID)`.
- `e.previewKind = PreviewKindTextFile`
- `e.previewPath = entry.Path`
- `e.previewTitle = entry.Text`
- Mark `Dirty = true`.

### 7. Extend Neovim adapter/port APIs

Files:

- `core/ports/ports.go`
- `neovim/nvim_adapter.go`

Likely add:

```go
SetBufferOption(bufNr int, key string, value any) error
DeleteBuffer(bufNr int, force bool) error
ExecLua(script string, result any) error // or narrower helpers
WindowCall(winId int, script string, result any) error // preferred if needed
```

Maybe also:

```go
SetWindowWidth(winId int, cols int) error
SetWindowHeight(winId int, rows int) error
```

Keep feature code behind port, not direct package globals.

### 8. Build text preview content from grid

File: `core/projection/file_explorer_projector.go`

Change `Project`:

```go
preview := p.buildPreview(state)
payload := map[string]any{
  "parent": p.buildPanePayload(parent, ...),
  "current": p.buildPanePayload(current, ...),
  "preview": preview,
  "currentWinMode": s.Mode,
}
```

`buildPreview` should switch on backend preview kind:

- directory: `kind: "directory"`, `title`, `directory: buildPanePayload(...)`
- local image: `kind: "localImage"`, `title`, `path`
- text file: find `previewWinID` grid, render via `renderContentPreview(grid)`

For text-file preview, pass real filetype if available from `WindowState.Filetype`. Current old code hardcoded `Filetype: "minifiles"`; fix that.

### 9. Resize preview Neovim window to match frontend content area

Files:

- `frontend/src/lib/file-explorer/FileExplorer.svelte`
- `app.go`
- `neovim/screen.go`
- possibly `features/file-explorer` or adapter

Frontend:

- Observe text preview grid content area below header.
- Send px width/height through `OnFileExplorerPreviewResize`.

Backend:

- Convert px to cells using same metrics as `CalculateGridSize` for now (`12x28`).
- Resize preview split/window to cols/rows.
- Keep focus in current pane after resize.

Need exact Neovim call:

- `nvim_win_set_width(previewWinID, cols)` for vertical split width.
- `nvim_win_set_height(previewWinID, rows)` if applicable.
- Or `SetWindowOption(win, "winfixwidth", true)` plus command resize.

### 10. Close/cleanup behavior

File: `features/file-explorer/file-explorer.go`

On close/reset:

- stop preview timer
- increment generation to invalidate pending callback
- delete/wipe preview scratch buffer if exists
- clear `previewKind`, `previewPath`, `previewTitle`
- preserve existing `resetSessionState`

### 11. Tests / verification

Manual:

1. Open file explorer.
2. Parent/current still render.
3. Move cursor:
   - selected row updates immediately
   - preview updates after ~75ms
4. Select directory:
   - preview shows read-only directory list
5. Select text file:
   - preview shows top of file
   - syntax highlights via filetype
   - current pane remains focused
6. Select image:
   - preview renders image
   - no base64 URL state
7. Select binary file:
   - placeholder, no raw bytes
8. Resize app / preview column:
   - text grid fills preview area below header
   - no cut-off rows/cols
9. Open same file modified elsewhere:
   - preview still shows disk snapshot, not modified buffer
10. Close explorer:
   - no stale preview callback, no buffer leaks obvious from `:ls`

Automated where practical:

- Go tests for preview kind selection by extension/stat.
- Go tests for binary detector.
- Go tests for read-line limit.
- Type check frontend after union changes.

## Risks / Follow-ups

- Go timer may mutate `FileExplorer` concurrently with Neovim event handlers/projector. Need verify runtime threading or serialize preview update work.
- Generated Wails bindings may need regeneration after `ReadLocalImage` comment/signature cleanup.
- Grid sizing still uses hardcoded `12x28` metrics. Good enough now; future should derive actual cell metrics from frontend CSS/font.
- Markdown image metadata path may change when `LocalImage` raw path handling changes. Need test markdown images too.
- Existing dead `buildPreview` helper in projector should be replaced or removed to avoid confusion.
