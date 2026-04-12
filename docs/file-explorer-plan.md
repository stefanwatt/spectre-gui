# File Explorer Feature Plan

## Architecture Context

The file explorer is built on top of **mini.files** (a patched fork at `/home/stefan/Projects/mini.files`). The data flow is:

1. **Neovim (mini.files)** — owns the buffer state, filesystem operations, and concealed path_id per line
2. **Go backend** — receives `MiniFilesBridge` RPC events, maintains a `Registry` of pane assignments (parent/current/preview win+buf), and projects parsed directory entries to the frontend via `file-explorer-update` events
3. **Svelte frontend** — receives projected state, renders the three-pane explorer UI, sends keypresses back through `SendKey` -> neovim

Key detail: each line in a mini.files buffer has a **concealed path_id** prefix that maps to the real filesystem path via an internal `H.path_index` table. Yank/delete on these lines preserves the path_id, and mini.files computes filesystem diffs from them on sync.

---

## Feature 1: Git Integration (Phased)

### Phase 1 — Git Status Indicators

**Goal:** Show git status (modified, staged, untracked, etc.) next to each file entry in all panes.

#### Path Resolution (Prerequisite)

`FileExplorerDirectoryEntry` has a `FilePath` field that is never populated. We need full paths to query git status.

- **Approach:** Query the directory path once per pane (the parent path of the buffer) and construct entry paths as `dir_path + "/" + entry.text`. The directory path can be obtained from `MiniFiles.get_fs_entry(buf_id, 1).path` and taking its dirname.
- **Alternative:** Extend the existing `GetMiniFilesDirectoryLineMap` Lua call to also return paths per line. This avoids separate RPC round-trips.

#### Git Status Query

- **New Go function:** `GetGitStatusForDirectory(dirPath string) map[string]string` — runs `git status --porcelain=v1 -z` and parses output into a map of `relativePath -> statusCode` (e.g. `"M "`, `"??"`, `"A "`, `" M"`, `"MM"`).
- **Caching:** Cache per-directory, invalidate on `MiniFilesBufferUpdate` events. Could also invalidate on a short timer or after stage/unstage operations.
- **Data flow:**
  - Extend `DirectoryEntry` with `GitStatus string` field
  - Projector populates it when building entries
  - Frontend renders a colored indicator next to the entry icon

#### Frontend Rendering

- Status indicator between icon and filename (small colored letter/dot)
- Color mapping: modified=yellow, staged=green, untracked=grey, conflict=red (standard conventions)
- Directories could show aggregated status ("worst" child status) — defer this if complex

#### Considerations

- `git status` on large repos: run once per explorer open with `--porcelain` for the full repo, then filter per-directory. Much faster than per-directory queries.
- Directory aggregation can be deferred to Phase 2.

### Phase 2 — Stage/Unstage from Explorer

- Keymaps: `s` to stage, `u` to unstage
- Backend: `git add` / `git reset HEAD` on the file path(s)
- Integrates with multi-select (Feature 2): if entries are selected, operate on all selected
- After operation: re-query git status and update frontend

---

## Feature 2: Multi-Select with Discontinuous Entry Operations

**Goal:** Select arbitrary (non-adjacent) entries, then yank or delete them as a batch.

### Selection State — Frontend (Svelte)

Selection lives in Svelte because neovim's visual mode only supports contiguous selections.

```typescript
// In state.svelte.ts
let _fileExplorerSelectedIds: Set<number> = $state(new Set());
```

Clear on: pane navigation, explorer close, buffer change. Only the "current" pane supports multi-select.

### Keymap: Space to Toggle Selection

- Intercept Space in `SendKey` when `keymapMode === "file-explorer"`. Don't forward to neovim.
- Emit `file-explorer-toggle-select` event with current cursor entry ID.
- After toggle, send `j` to neovim to advance cursor (rapid multi-select by holding Space).
- Need to verify mini.files doesn't use Space for anything.

### Visual Indicator

- Selected entries: distinct background color (separate from cursor highlight)
- **Status bar** at bottom of explorer: `"3 entries selected"` + action hints (`y yank  d delete  Esc clear`)
- Only appears when `selectedIds.size > 0`

### Yank Operation (`y` with selections active)

1. Frontend sends `file-explorer-yank` event to Go backend with selected entry IDs (row indices)
2. Go backend calls neovim Lua:
   - `vim.api.nvim_buf_get_lines(buf_id, row, row+1, false)` for each selected row
   - `vim.fn.setreg('"', lines, 'l')` — puts them linewise in the unnamed register
3. Frontend clears selections
4. User can now paste (`p`) in mini.files — concealed path_ids tell mini.files what to copy

### Delete Operation (`d` with selections active)

1. Frontend sends `file-explorer-delete` event to Go backend with selected entry IDs
2. Go backend calls neovim Lua:
   - Also yank the lines into register first (like vim's `dd`)
   - Sort row indices descending (avoid index shifting)
   - `vim.api.nvim_buf_set_lines(buf_id, row, row+1, false, {})` for each
3. Frontend clears selections
4. Buffer modification → mini.files sees entries as "deleted" on next sync (`=`)
5. Neovim's `u` (undo) restores deleted lines naturally

### Edge Cases

- Navigate to subdirectory while entries selected → clear selections
- Selected entry is also cursor entry → both highlights shown, works normally
- Undo after delete → neovim undo restores buffer lines

---

## Feature 3: Keymap Hints Floating Window

**Goal:** A reusable floating component showing available keymaps for the current context. Usable in file explorer, pickers, and future features.

### Reusable Component: `KeymapHints.svelte`

```typescript
interface KeymapHintsProps {
  hints: { key: string; description: string }[];
  position?: 'bottom-right' | 'bottom-left' | 'bottom-center';
}
```

- Small floating panel in the corner of the parent container
- Semi-transparent background, compact two-column layout: key (styled as `<kbd>`) | description
- Auto-hides if hints array is empty

### File Explorer Hints

Contextual: when entries are selected show action hints (y/d/Esc), otherwise show navigation hints (h/l/j/k/=/q). Keeps the panel small and relevant.

### Picker Hints

Each picker provides its own hint set via the same component.

### Toggle

`?` toggles hints on/off (preference persisted in localStorage). Visible by default for discoverability.

---

## Feature 4: Image Preview

**Goal:** Show actual images instead of `-Non-text-file---` in the preview pane.

### Detection (Frontend-side)

Check if selected entry's filename ends with a known image extension (`.png`, `.jpg`, `.jpeg`, `.gif`, `.webp`, `.svg`, `.bmp`, `.ico`).

### Rendering

- When preview is content type (not directory) AND selected entry is an image:
  - Render `<img src="/local-image/{encodeURIComponent(fullPath)}" />` instead of `<Grid>`
  - Uses the existing `/local-image/` file server (already serves images for markdown)
  - `object-fit: contain`, dark background to match explorer theme

### Path Construction

Requires the full path of the selected entry — ties into Feature 1's path resolution prerequisite. Can be constructed as `directory_path + "/" + entry.text` if we add the directory path to the explorer state.

### Fallback

If image fails to load, fall back to existing grid content. `onerror` handler on the `<img>` tag.

---

## Feature 5: Simplify and Internalize mini.files

**Goal:** Reduce the patched mini.files fork to the bare minimum needed for the nvim-gui file explorer, then migrate as much logic as possible into the Go codebase. Ideally eliminate the external Lua dependency entirely (or reduce it to a tiny inlined snippet).

### Why

mini.files is a general-purpose TUI file explorer with a lot of complexity for things we don't need:
- Dynamic column count / window layout — we always use a fixed 3-column layout (parent, current, preview)
- Complex window management (positioning, resizing, border rendering) — the GUI renders its own UI, mini.files windows are invisible to the user
- Filter/sort/prefix customization — we use a fixed configuration
- Help/mapping system — we handle keymaps on the frontend
- Content manipulation UX (visual cues, statusline, etc.) — all rendered by Svelte

### Approach (Incremental)

#### Phase 1 — Audit and strip unused mini.files features

- Go through the patched `mini/files.lua` and identify which functions/features are actually called or depended on
- Remove unused features: custom filters, custom sort, custom prefix, help system, statusline integration, any TUI-specific rendering logic
- Keep: buffer management with concealed path_ids, `get_fs_entry()`, filesystem diff computation (`H.buffer_compute_fs_diff`), synchronize/apply logic, the autocmd event bridge (`MiniFilesBridge`)

#### Phase 2 — Move logic to Go where possible

Things currently done in Lua that could move to Go:
- **Directory listing and entry construction:** Instead of mini.files scanning the filesystem and populating buffers, Go could scan directories (using `os.ReadDir`), build the entry list, and write buffer contents via the nvim API. This is faster and gives us direct access to file metadata (size, permissions, git status) without extra RPC calls.
- **Path resolution:** Instead of querying `get_fs_entry()` via Lua RPC, Go owns the path index directly.
- **Preview content:** Go already handles preview rendering through the projector. If Go also owns directory scanning, it can decide preview content without mini.files involvement.
- **Icon resolution:** Currently mini.files calls `MiniIcons.get()` or nvim-web-devicons. Go could do icon lookup directly (it already has icon data for the picker).

Things that likely **must stay in Lua/neovim:**
- Buffer creation and management (neovim buffers are inherently nvim objects)
- Concealed text / extmarks for path_ids (neovim rendering feature)
- The filesystem diff computation and apply logic (tightly coupled to buffer state)
- Undo/redo (neovim's undo tree)

#### Phase 3 — Inline remaining Lua

Once mini.files is stripped down, the remaining Lua should be small enough to:
- Either inline directly in the Go codebase as `ExecLua` strings (if truly minimal, <50 lines)
- Or keep as a single small `.lua` file shipped with the app (not an external fork dependency)

This eliminates the dependency on `/home/stefan/Projects/mini.files` and the `EnsureMiniFilesPatched()` hot-loading mechanism.

### End State

- No external mini.files fork dependency
- File explorer buffer management is a thin Lua layer (buffer creation, concealed path_ids, undo support)
- Directory scanning, entry building, path indexing, icon resolution, git status — all in Go
- Filesystem operations (rename, delete, copy/move) computed in Go by diffing buffer state against the known directory contents, applied via Go's `os` package (with confirm prompts going through the existing frontend flow)

---

## Implementation Order

Based on dependencies and incremental value:

1. **Image Preview** (Feature 4) — quickest win, minimal changes
2. **Path Resolution** (Feature 1 prerequisite) — needed for git status, makes image preview robust
3. **Keymap Hints Component** (Feature 3) — standalone, immediately useful
4. **Git Status Indicators** (Feature 1 Phase 1) — builds on path resolution
5. **Multi-Select** (Feature 2) — most complex, benefits from keymap hints being available
6. **Git Stage/Unstage** (Feature 1 Phase 2) — builds on multi-select + git status
7. **Strip mini.files** (Feature 5 Phase 1) — audit and remove unused code from the fork
8. **Migrate to Go** (Feature 5 Phase 2-3) — move directory scanning, path indexing, icons to Go; inline remaining Lua

---

## Open Questions

- [ ] Should multi-select use Space or another key? Need to verify mini.files doesn't bind Space.
- [ ] Should git status query the entire repo at once or per-directory? (Per-repo is faster)
- [ ] How should directory git status aggregation work? Worst child status? Count badge?
- [ ] Should keymap hints be toggleable per-feature or globally?
- [ ] For image preview: support transformations (resize for thumbnails) or just native rendering?
