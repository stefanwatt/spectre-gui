# File Explorer Backend Cursor Column Plan

## Goal

Move file-explorer cursor column logic from Svelte into Go.

Fix:
- wrong cursor column for pending/draft rows with no `id/` prefix
- cursor placement too far left before file/dir name
- programmatic pane sync always setting column `0`

Target behavior:
- `cursorCol` in UI payload means **visible 0-based column inside `entry.text`**
- frontend does **no hidden-prefix math**
- cursor cannot go left of filename start
- vertical movement preserves visible column like normal file navigation, clamped to target name length

---

## Current Problem

Current split of responsibility:

- `features/file-explorer/file-explorer.go`
  - `UpdateSelectedyEntry()` stores raw Nvim buffer `col`
  - `syncPaneCursorToSelection()` restores row with `col = 0`
- `frontend/src/lib/file-explorer/FileExplorer.svelte`
  - subtracts `String(entry.id).length + 1`
  - assumes every line is rendered from hidden `"<id>/"` prefix

This breaks for pending/draft rows because raw current buffer line may not have prefix at all.

Extra mismatch:
- Nvim file-explorer buffer does **not** conceal prefix
- raw buffer columns therefore vary with prefix width
- frontend hides prefix anyway
- preserving raw buffer column does **not** preserve visible filename column when id widths differ

So backend must translate between:
- raw buffer column
- visible filename column

---

## Shared Decisions

### 1. Backend owns cursor semantics

Go becomes source of truth for file-explorer cursor column.

Frontend only renders already-normalized `cursorCol`.

### 2. `Directory.CursorCol` changes meaning

`Directory.CursorCol` stores:
- visible 0-based column inside selected entry text
- clamped to `[0, len(visible text)]`

Not raw Nvim buffer column.

Applies to parent/current pane payloads.

### 3. Prefix detection uses raw buffer line

Cursor math uses actual current buffer line text, not `entry.ID` assumptions.

Reason:
- pending rows may not have id prefix
- current buffer can temporarily diverge from model during edits
- cursor should follow what user is editing, not what model hopes line looks like

### 4. Prefix parser rule

Treat hidden prefix only as:
- one or more leading digits
- immediately followed by `/`

Regex shape:
- `^[0-9]+/`

Examples:
- `12/foo` -> prefix width `3`, visible text `foo`
- `foo` -> prefix width `0`
- `123` -> prefix width `0`
- `foo/bar` -> prefix width `0`

### 5. No trimming for cursor math

Use raw line bytes/string.

Do **not** `TrimSpace()` for cursor-position calculations.

Reason:
- leading/trailing spaces in draft names are real editable content
- trimming shifts cursor incorrectly

### 6. mini.files-like left bound

For manual movement in current pane:
- cursor may stay anywhere inside visible name
- if it moves left of visible name start, backend pushes it right to filename start
- backend does not force snap to start on every move

### 7. Vertical movement preserves visible column

When selection changes to different row:
- preserve prior visible `CursorCol`
- reproject it onto destination row
- clamp to destination visible text length

Reason:
- raw Nvim vertical movement preserves raw buffer column
- raw column alone is wrong for this explorer because hidden prefix widths vary

### 8. Programmatic sync should set sane target, not blindly `0`

`syncPaneCursorToSelection()` should place cursor at:
- destination row
- mapped raw buffer column for stored visible `CursorCol`

This is not extra enforcement step.
Just correct target selection restore.

### 9. Existing autocmd path is enough

No new autocmd design needed.

Current `CreateBufferAutocmd()` already gives `CursorMoved` and `CursorMovedI` for current pane.
Use that path for backend correction.

---

## Proposed Backend Helpers

Add small pure helpers in `features/file-explorer/file-explorer.go`.

Suggested shape:

- `bufferLineVisibleStartCol(line string) int`
  - returns raw buffer column where visible filename starts
- `bufferLineVisibleText(line string) string`
  - strips only recognized leading `^[0-9]+/`
- `bufferColToVisibleCol(line string, rawCol int) int`
  - maps raw buffer column to visible text column
  - clamps to `[0, len(visible text)]`
- `visibleColToBufferCol(line string, visibleCol int) int`
  - maps visible text column back to raw buffer column
  - clamps to visible text length first

Keep helpers based on raw buffer line only.
No `entry.ID` dependency.

---

## Implementation Plan

## Slice 1 — Normalize cursor model in Go

### Files
- `features/file-explorer/file-explorer.go`
- maybe `features/file-explorer/event-handlers.go`
- tests in `features/file-explorer/*_test.go`

### Changes

Refactor `UpdateSelectedyEntry()` so it no longer stores raw `col`.

New behavior:
1. read current raw buffer line for `row`
2. compute selected entry id from row as today
3. map raw buffer column into visible column
4. store visible column in `e.current.CursorCol`
5. if cursor is left of visible start, compute corrected raw target and push Nvim cursor there

Need reentry-safe behavior:
- only call `SetWindowCursor()` when corrected raw column differs from actual raw column
- guard against cursor-write loop from `CursorMoved`

Recommended guard options:
- small boolean reentry flag on `FileExplorer`, or
- last-corrected `(row,col)` check

Result after this slice:
- `cursorCol` payload already means visible column
- pending rows with no prefix report correct visible column
- moving left into prefix gets corrected in backend

### Important detail: row changes

When `CursorMoved` lands on different selected row:
- preserve previous visible `e.current.CursorCol`
- compute destination raw column from new line + previous visible col
- if current raw column differs, correct Nvim cursor to mapped raw target
- store preserved/clamped visible col

Without this, visible column still drifts when id widths differ.

---

## Slice 2 — Fix programmatic selection restore

### Files
- `features/file-explorer/file-explorer.go`
- tests in `features/file-explorer/sync_test.go`

### Changes

Refactor `syncPaneCursorToSelection()`.

Current behavior:
- finds selected row
- restores cursor with `col = 0`

New behavior:
1. find selected row
2. read actual raw buffer line at that row
3. map stored visible `directory.CursorCol` to raw buffer column
4. call `SetWindowCursor(winID, row, rawTargetCol)`

Rules:
- if `directory.CursorCol == 0`, raw target becomes visible name start
- if buffer line has no recognized prefix, raw target equals visible column
- clamp to visible text length

This should apply to parent and current panes.

---

## Slice 3 — Remove prefix math from Svelte

### Files
- `frontend/src/lib/file-explorer/FileExplorer.svelte`

### Changes

Remove hidden-prefix subtraction:
- delete `currentCursorDisplayCol()` logic based on `entry.id`
- make split/render logic use payload `cursorCol` directly

Frontend may keep tiny defensive clamp to `entry.text.length` before `slice()`.
That is render safety, not prefix calculation.

Desired frontend contract:
- `cursorCol` already points inside visible `entry.text`
- no knowledge of `id/` format
- no special pending/draft branch

---

## Slice 4 — Tests

Add/update tests for both pure helpers and integration-ish cursor behavior.

### Pure helper cases

`bufferLineVisibleStartCol()` / mapping helpers:
- `"12/foo"` -> start `3`
- `"foo"` -> start `0`
- `"123"` -> start `0`
- `"foo/bar"` -> start `0`
- `" 12/foo"` -> start `0`
- `"12/ foo"` -> visible text starts with leading space after slash
- `"12/"` -> empty visible text, start after slash

### `UpdateSelectedyEntry()` cases

- stores visible col for prefixed line
- stores visible col for no-prefix pending line
- clamps left-of-name move back to visible `0`
- preserves visible col when moving to row with different prefix width
- clamps preserved visible col to shorter destination name length

### `syncPaneCursorToSelection()` cases

Update existing expectation in `features/file-explorer/sync_test.go`.

Current expectation:
- cursor restore uses `col: 0`

New expectation:
- cursor restore uses raw target column matching filename start or preserved visible column

Need at least:
- prefixed row with visible `0` -> raw start column
- prefixed row with visible `3` -> raw start + `3`
- no-prefix row with visible `2` -> raw `2`

### Frontend behavior cases

If frontend tests exist or are easy to add:
- pending row renders cursor at payload `cursorCol`
- no `id`-length dependency remains

---

## Notes About mini.files

mini.files uses backend-side cursor correction on explicit cursor set and on cursor-move events.
It only needs lower-bound tweak because its hidden prefix is handled in actual Nvim view.

Our explorer differs:
- prefix is hidden only in frontend
- Nvim buffer still uses raw `id/text`
- raw column therefore cannot be forwarded directly

So matching mini.files behavior here means:
- same left-bound invariant
- plus explicit visible/raw column translation in Go

Not exact same implementation.

---

## Out of Scope

Not part of this slice:
- changing visible line format in buffer
- adding Nvim conceal for id prefix
- unicode grapheme / double-width exact cursor semantics
- broader draft-entry identity redesign
- renaming `UpdateSelectedyEntry()` typo unless touched opportunistically

---

## Acceptance Criteria

Done when all true:

- pending/draft row with no id prefix shows correct cursor column in frontend
- Svelte no longer subtracts `id` prefix width
- moving left cannot place cursor before visible filename start
- moving up/down preserves visible filename column across different id widths
- programmatic pane refresh restores cursor to filename-relative column, not raw `0`
- existing selection/pane sync behavior still works
- tests cover prefix/no-prefix rows and restore behavior
