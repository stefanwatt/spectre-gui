# File Explorer Live Buffer Sync Plan

## Goal

Keep `FileExplorer.current.Entries` in sync with live edits in current file-explorer buffer.

Scope for this plan:
- current pane only
- live text reflection on every keystroke
- no filesystem sync yet
- no icon recovery work yet
- no diff/rename/delete semantics beyond what UI needs right now

Out of scope for this slice:
- applying edits to filesystem
- conflict resolution on sync
- exact duplicate handling on sync
- icon/path reconciliation after destructive edits
- richer preview semantics after identity loss

---

## Why `AttachBuffer`

Current architecture has two different state streams:

1. **Grid redraw** in `core/reducer/screen.go`
   - good for visual rendering
   - not semantic source for file-explorer entries
2. **`FileExplorer.current.Entries`**
   - this is what frontend currently renders for parent/current panes
   - must reflect edits live

Because `Entries` remain source of truth, current pane needs direct buffer-update feed.

Decision:
- use `nvim_buf_attach` / `AttachBuffer` for current pane buffer only
- keep grid redraw pipeline unchanged for visual concerns
- do not try to derive current pane semantic state from redraw events for this slice

---

## Behavioral Model

### Visible buffer format stays as-is

Current pane buffer line format stays:

`<hidden-or-visible-id>/<text>`

Reason:
- yank/delete workflows depend on line content preserving mini.files-like identity behavior
- user explicitly wants same behavior style as mini.files

### Identity loss is acceptable

If user destroys id prefix, identity is considered lost.

Examples:

#### Case A: inline rename preserving id
- line starts as `123/foo.txt`
- user changes visible text to `123/bar.txt`
- entry still associated with prior item because id still parseable
- preview can still follow old association for now

#### Case B: destructive edit breaking id
- line starts as `123/foo.txt`
- user presses `cc`
- line no longer has parseable id
- row remains visible
- old association is gone
- row becomes pending/new-like entry from perspective of future sync

This is expected behavior, not bug.

---

## Vertical Slices

## Slice 1 — Attach current buffer, prove event pipe

### Goal

Receive reliable update signal whenever current file-explorer buffer changes.

### Changes

- extend `ports.NvimClient` with buffer-attach capability
- implement adapter support in `neovim/nvim_adapter.go`
- attach only to `e.current.BufNr`
- register handler for buffer update notifications in `neovim.go`
- gate handling on:
  - file explorer active
  - event buffer == current file-explorer buffer

### Non-goals

- no entry mutation yet
- no sync logic
- no preview logic changes

### Acceptance criteria

- typing in current pane fires handler
- `Shift-C`, `cc`, `dd`, `o`, `O`, `p`, `P` all fire handler
- parent and preview panes are not attached
- no regressions in existing cursor-move handler

### Verification

Manual only for this slice:
- open explorer
- enter insert/normal edits in current pane
- verify handler logs or observable update hook fires every time buffer changes

---

## Slice 2 — Rebuild `current.Entries` from whole-buffer snapshot

### Goal

Mirror current buffer contents into `e.current.Entries` live.

### Core decision

On each relevant buffer update event, re-read full current buffer and rebuild `e.current.Entries` from snapshot.

Reason:
- buffer size small
- implementation simpler
- handles all edit shapes uniformly
- avoids fragile incremental splice logic too early

### Parsing rules

For each line in current buffer:

#### If line has parseable id prefix
- preserve parsed id as entry identity
- extract text after first `/`
- update `entry.Text`

#### If line has no parseable id prefix
- treat whole line as visible text
- create entry representation that can still render stably in frontend
- consider row semantically detached from previous known file identity

### Important rendering requirement

Broken-id rows still need stable frontend keys.

Do **not** reuse one constant fallback id like `0` for all broken lines.

Need stable per-row identity for rendering, otherwise keyed Svelte list will churn badly on edits.

Exact internal mechanism can be chosen during implementation.
Possible options:
- generated fallback render ids
- separate internal render id distinct from parsed buffer id
- row-stable pending ids reused across rebuilds where possible

Plan does **not** lock exact representation yet.

### Acceptance criteria

- rename with intact id updates frontend text live
- `cc` shows empty row in frontend, no collapse
- `o` / `O` creates new visible row in frontend as user types
- `dd` removes deleted line from frontend because buffer line is gone
- line order exactly follows buffer order
- no resorting done during editing

### Verification

Manual scenarios:

1. rename preserving id
   - `Shift-C` on line
   - edit filename
   - frontend mirrors each keystroke

2. clear line breaking id
   - `cc`
   - frontend shows empty row
   - no panic

3. create line
   - `o`
   - type text
   - new row appears live

4. delete line
   - `dd`
   - row disappears from frontend immediately

5. paste lines
   - `p` / `P`
   - frontend reflects inserted rows immediately

---

## Slice 3 — Fail-soft selection when current line id is broken

### Goal

Stop treating broken ids in current pane as fatal during cursor updates.

### Problem

Current selection logic parses id from current buffer line. If parse fails, current code path errors.
That is incompatible with accepted behavior where destructive edits may intentionally destroy id.

### Desired behavior

If selected current line no longer has parseable id:
- keep cursor column updated
- do not panic
- clear semantic selection association for current pane
- mark file explorer dirty so frontend can react

### Notes

This slice only makes current behavior safe.
It does not solve future sync semantics.

### Acceptance criteria

- move cursor onto broken-id line
- app stays stable
- no panic from `UpdateSelectedyEntryCurrent`
- stale old selection association is not kept accidentally

---

## Slice 4 — Optional hardening after basic live sync works

Only do after slices 1-3 verified.

Possible follow-ups:
- handle detach/reload events
- avoid redundant full-buffer rebuilds when only `changedtick` arrives without text changes
- add tests around line parsing and rebuild behavior
- improve stability of fallback identities for broken-id lines
- decide preview policy when selected line loses identity

---

## Data/State Notes

### Current source of truth

For parent/current pane rendering, source of truth remains `FileExplorer` state, not grid cells.

### Parent pane

Parent pane stays non-editable and unchanged in this plan.

### Preview pane

Preview behavior after id loss is deferred.
Current plan only ensures entry text reflects live buffer edits.

### Sorting

No sorting while editing.
Buffer order is authoritative until future sync/reload step.

---

## Risks

### Duplicate semantic identities

Pasting copied explorer lines may duplicate visible ids or filenames.
This plan does not solve future sync behavior around duplicates.
For now, UI must only mirror buffer state faithfully.

### Broken-id rows need stable render identity

If implementation uses unstable fallback identity, frontend may jitter or recreate DOM nodes too often.
This matters especially for empty lines and rapid insert-mode edits.

### Selection semantics become partial

When id breaks, current pane selection can no longer point to known underlying entry.
That is acceptable for now.
Need future explicit model for pending/new rows.

---

## Decisions captured

- `Entries` stay source of truth
- current pane only uses buffer attach
- live updates happen on each keystroke
- hidden id remains in buffer text
- broken id is acceptable and expected
- no resort during editing
- frontend stays dumb
- whole-buffer rebuild preferred before any incremental diffing

---

## Open questions for later

- exact internal representation for broken-id rows
- whether `DirectoryEntry` should gain draft/original metadata
- how preview should behave for detached rows
- how sync should reconcile duplicates and copies
- whether parent pane selection model should later align more explicitly with current pane edge cases
