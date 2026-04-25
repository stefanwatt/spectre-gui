# File Explorer Filesystem Sync — Vertical Slice Plan

## Goal

Add explicit filesystem synchronization for file-explorer drafts.

User edits explorer buffers first. Disk changes happen only on `=`.

Must support cross-directory workflow in one session:

1. `dd` file in child dir
2. `<Left>` go to parent
3. `p` paste into parent
4. `=` sync

Must support revisit unsynced dir:

1. create `foo.txt` line
2. go out
3. go in again
4. unsynced `foo.txt` still visible

---

## Decisions locked from Q&A

- Sync model: **explicit apply** (not live disk writes)
- Apply key: **`=`**
- Quit dirty explorer: show prompt
- `<Left>`: **no dirty prompt**
- Keep one buffer per directory (no buffer reuse)
- Keep visited dir buffers for whole explorer session (until close)
- No memory optimization/eviction in this slice
- Dirty tracking by map presence (no extra `Dirty bool` in draft record)
- Sync scope: **all dirty dirs** in session
- Dirty registry key: **bufnr**
- Add immutable global disk-source index: `sourceByID`
- `sourceByID` populated from real filesystem entries only (immutable during session)
- New directory create syntax: trailing `/` means directory, else file
- Error strategy: minimal, no full preflight, fail-fast on first apply error
- Keep incremental line-update strategy (no full-buffer rebuild)

---

## Why architecture change needed

Current design reuses one current buffer and swaps directory data.

Problems:

- Unsynced edits disappear when leaving/re-entering dir
- `nvim_buf_lines_event` ranges can hit wrong in-memory slice after nav and panic
  - observed: `panic: runtime error: slice bounds out of range [168:11]`

One-buffer-per-dir removes mismatch class and preserves drafts naturally.

---

## Scope

### In scope

- Per-directory buffer persistence across navigation
- Per-buffer dirty draft baseline capture
- Global sync (`=`) over all dirty dirs
- Minimal fs action derivation: create/delete/copy/move/rename
- Minimal apply pipeline with fail-fast errors
- Quit prompt for dirty explorer: Sync / Discard / Cancel

### Out of scope

- Advanced conflict UX / bulk conflict resolver
- Rollback/transaction semantics
- Memory eviction for dir buffers
- Fancy preflight validation
- Filesystem watcher reconciliation during sync

---

## Target runtime model

### 1) Directory buffer cache

Keep per-dir runtime objects for session:

- `dirByBuf map[int]*Directory`
- `bufByPath map[string]int` (or equivalent helper)

Each visited dir gets dedicated buffer once.

Revisit dir => reuse same buffer + same unsynced text.

### 2) Dirty draft registry

Dirty dirs tracked by bufnr:

- `dirtyByBuf map[int]*DirDraft`

`DirDraft` minimal:

- `BufNr int`
- `DirPath string`
- `OriginalEntries map[uint64]DirectoryEntry` (frozen first dirty event)

No stored `WorkingLines`.

At sync, read live lines from buffer directly.

### 3) Immutable source identity index

Global session map:

- `sourceByID map[uint64]DirectoryEntry`

Filled only from filesystem-loaded entries.

Never overwritten by draft-generated IDs.

Purpose: infer `from` path for parseable hidden ids across directories.

---

## Sync diff model

At `=`:

1. Iterate all `dirtyByBuf`
2. Read full current lines for each dirty buffer via `GetBufferLines`
3. Build per-line diff intents:
   - parse line as `<id>/<text>`
   - blank/whitespace-only line: ignore
   - `to` path = `DirPath + name`
   - if line ends with `/`, `to` treated as directory target
   - `from` logic:
     - parseable `id` and `id` in `sourceByID` => `from = sourceByID[id].Path`
     - else `from = nil` (new/create-like)
4. Missing originals:
   - For any `id` in `OriginalEntries` not present in current parsed-id set => `{from: original.Path, to:nil}` delete candidate

Then classify actions (mini.files style):

- `from=nil,to!=nil` => create
- `from!=nil,to=nil` => delete
- `from!=nil,to!=nil` => raw copy candidate
- collapse `delete+copy` of same `from`:
  - same parent dir => rename
  - different parent dir => move
- remaining raw copy => copy

---

## Apply model

Order:

1. copy
2. create
3. move
4. rename
5. delete

Rules:

- No full preflight pass
- First filesystem error => stop immediately
- Keep explorer open
- Keep dirty drafts untouched for retry/fix
- Surface failed op + error message

On full success:

- clear `dirtyByBuf`
- refresh affected dirs from disk
- rebuild pane views while preserving navigation context

---

## Navigation/interaction semantics

### `GoIn` / `GoOut`

- Must switch pane windows to cached dir buffers, not recreate/overwrite reusable current buffer model
- `<Left>` never prompts, even if dirty

### `q` / close

If no dirty dirs: close immediately.

If dirty dirs exist: prompt with choices:

1. Sync
2. Discard
3. Cancel

- Sync => run global `=` flow, close only if success
- Discard => drop drafts/dirty map and close
- Cancel => do nothing

### `=` / synchronize

- Available as current pane buffer-local keymap
- Runs global sync on all dirty dirs
- No extra confirm dialog required in this slice

---

## Event/attachment rules

- Attach to every newly created directory buffer: `AttachBuffer(buf, false, ...)`
- Route `nvim_buf_lines_event` by `bufnr` into matching cached directory
- Mark dir dirty lazily on first external edit event for that buffer
- Freeze `OriginalEntries` on first dirty event only

Important hardening with incremental splice kept:

- Guard `firstline/lastline` before any slice operation in `UpdateEntries`
- Reject or clamp invalid ranges to avoid panic

---

## Implementation slices

## Slice A — Buffer-per-dir foundation + panic removal

### Goal

Replace reusable-current-buffer model with persistent per-dir buffers.

### Changes

- Introduce dir buffer cache maps in `FileExplorer`
- Refactor `Open`, `GoIn`, `GoOut` to switch to cached buffers
- Ensure keymaps/autocmd applied for current active dir buffer
- Attach every new dir buffer
- Route buffer-line events by bufnr
- Add bounds guard in incremental `UpdateEntries`

### Acceptance

- Re-entering dir shows unsynced edits unchanged
- No panic on rapid nav + edits
- `nvim_buf_lines_event` updates correct dir state by bufnr

## Slice B — Dirty draft baseline + global source index

### Goal

Track dirty dirs and frozen originals for sync.

### Changes

- Add `dirtyByBuf`
- Capture `OriginalEntries` on first dirty event per bufnr
- Add immutable `sourceByID` from disk entries
- Add `=` keymap and RPC handler stub to compute/print actions (no fs writes yet)

### Acceptance

- `=` produces deterministic action plan across multiple dirty dirs
- Cross-dir `dd` + `<Left>` + `p` produces move/copy-style intents

## Slice C — Filesystem apply + close prompt

### Goal

Execute action plan on disk and handle dirty-close UX.

### Changes

- Implement apply ops in defined order
- Fail-fast error handling
- Wire quit prompt Sync/Discard/Cancel
- Keep `<Left>` ungated
- On success, clear dirty state and refresh

### Acceptance

- Required sequence works end-to-end (`dd` child, `<Left>`, `p`, `=`)
- `q` on dirty shows prompt and follows chosen branch
- `<Left>` always navigates without prompt

---

## Manual test matrix

1. **Revisit unsynced dir**
   - Add `foo.txt` line, go out, go in
   - Expect `foo.txt` still present

2. **Cross-dir move intent**
   - Child: `dd` item
   - Parent: `p`
   - `=`
   - Expect file moved/copied+deleted per intent

3. **Cross-dir copy intent**
   - Duplicate line in parent without deleting source
   - `=`
   - Expect copy action

4. **Rename same dir**
   - Edit text with parseable id preserved
   - `=`
   - Expect rename

5. **Create file/dir**
   - New `a.txt`, new `dir/`
   - `=`
   - Expect file + directory creation

6. **Dirty quit prompt**
   - Edit then `q`
   - Expect Sync/Discard/Cancel prompt behavior

7. **No left prompt**
   - Dirty + `<Left>`
   - Expect navigation without prompt

8. **Error fail-fast**
   - Force conflict (target exists)
   - `=`
   - Expect stop at first error, drafts preserved

---

## Files likely touched

- `features/file-explorer/file-explorer.go`
- `features/file-explorer/event-handlers.go`
- `core/ports/ports.go` (only if extra nvim ops needed)
- `neovim/nvim_adapter.go` (only if extra nvim ops needed)
- `app.go` (confirm choice plumbing)
- `frontend/src/lib/runtime-events-service.ts` (prompt flow verify)
- `frontend/src/lib/file-explorer/FileExplorerConfirmPrompt.svelte` (reuse as-is if possible)

---

## Notes

Keep code minimal.

No overengineering for generic file manager behavior.

Primary target: fast project-local ops with reliable draft persistence + explicit sync.