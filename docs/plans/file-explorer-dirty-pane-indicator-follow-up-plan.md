# File Explorer Dirty Pane Indicator — Follow-up Plan

## Goal

Show per-pane dirty indicator in file explorer UI.

Indicator meaning: pane has unsynced file-explorer draft edits (same truth as `dirtyByBuf`).

Visual target: pane right border turns yellow.

---

## Decisions locked

- Dirty semantic: **unsynced file-explorer edits only**.
- Data source: **backend `dirtyByBuf` only**.
- Dirty contract: `dirty` is **required** on `App.FileExplorerDirectory`.
- Injection point: projector reads `FileExplorer.IsBufDirty(bufNr)`.
- Preview dirty: **deferred** until preview payload contract refactor.
  - Current slice only marks `parent` + `current` panes.

---

## Why this is needed

Filesystem sync slice done. Drafts persist and sync works, but UI lacks per-pane dirty signal.

Without indicator:
- user cannot see which visible dir still has unsynced edits
- quit/sync prompt appears "late" relative to visual workflow

---

## Scope

### In scope

- Backend query API for dirty-by-buffer.
- Projector enriches `parent/current` payload with `dirty` boolean.
- Frontend type update (`dirty` required).
- Pane border style update in file explorer component.

### Out of scope

- Preview dirty indicator.
- Preview payload union refactor (`preview.directory` vs plain `preview`).
- Any change to sync algorithm/apply behavior.

---

## Design

### 1) Backend dirty query API

File: `features/file-explorer/file-explorer.go`

Add minimal method:

```go
func (e *FileExplorer) IsBufDirty(bufNr int) bool
```

Behavior:
- return `false` for `bufNr <= 0`
- return map presence in `dirtyByBuf`

No duplication into `Directory` model state.

### 2) Projector payload enrichment

File: `core/projection/file_explorer_projector.go`

Current payload emits plain directories.

Change projection payload for `parent/current` to include `dirty` field from `IsBufDirty`:
- `parent.dirty = IsBufDirty(parent.BufNr)`
- `current.dirty = IsBufDirty(current.BufNr)`

Implementation detail:
- keep `fileexplorer.Directory` domain struct unchanged
- build view DTO / map in projector layer

Preview remains unchanged in this slice.

### 3) Frontend type contract

File: `frontend/src/app.d.ts`

Update `App.FileExplorerDirectory`:
- add required `dirty: boolean`

### 4) Frontend rendering

File: `frontend/src/lib/file-explorer/FileExplorer.svelte`

In pane wrapper snippet, apply dirty class when `data.dirty` true.

Base today:
- `border-r-2 border-r-surface0`

Add override class (example):
- `class:dirty-pane={data.dirty}`

Add CSS rule with higher priority to set yellow border color.
Use theme token path consistent with Catppuccin/daisy theme warning lane.

---

## Implementation slices

## Slice 1 — Backend + projector

### Goal

Expose dirty-by-buffer signal and include in `file-explorer-update` payload.

### Changes

- Add `IsBufDirty(bufNr int) bool` on `FileExplorer`.
- Update projector payload for `parent/current` to include required `dirty`.

### Acceptance

- editing current explorer buffer flips `current.dirty=true` in emitted payload.
- after successful sync, emitted payload has `dirty=false` for visible panes.

## Slice 2 — Frontend contract + UI

### Goal

Render yellow border for dirty panes.

### Changes

- Update `App.FileExplorerDirectory` type.
- Add dirty class binding in pane snippet.
- Add CSS border-color override for dirty pane.

### Acceptance

- dirty pane border visibly yellow.
- clean pane border remains `surface0`.
- no runtime TS errors from missing `dirty`.

---

## Manual test matrix

1. **Current pane dirty on edit**
   - Open explorer, edit one entry in current pane.
   - Expect current pane yellow border.

2. **Dirty clears on sync success**
   - Edit pane, press `=`.
   - Expect border returns to default after success refresh.

3. **Dirty preserved across navigation**
   - Edit dir A, navigate to parent/child, return.
   - Expect pane for dirty dir still yellow when visible.

4. **Multiple dirty dirs visible**
   - Make drafts in two dirs, render both parent/current as dirty.
   - Expect both visible dirty panes yellow.

5. **Discard close branch**
   - Dirty edits, press `q`, choose Discard.
   - Explorer closes without extra update; reopen starts clean (no yellow).

6. **No behavior regression**
   - Existing sync/apply actions unchanged.
   - `<Left>` still ungated by dirty prompt.

---

## Files likely touched

- `features/file-explorer/file-explorer.go`
- `core/projection/file_explorer_projector.go`
- `frontend/src/app.d.ts`
- `frontend/src/lib/file-explorer/FileExplorer.svelte`

---

## Dependency note

Preview dirty indicator blocked by known preview payload contract mismatch.
Tracked in:
- `docs/plans/file-explorer-filesystem-sync-vertical-slice-plan.md` (`## Notes`)
