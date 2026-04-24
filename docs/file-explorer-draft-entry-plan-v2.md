# File Explorer Pending Entry Plan v2 (Buffer-First Incremental)

## Why v2

Your comments changed direction on 2 points:
- full-buffer rebuild every change = unnecessary
- helper parse model can be simpler

So v2 keeps original goal (`o` shows new row fast) but changes sync strategy.

---

## Design Writeup: Current Architecture Tradeoff

## Current design (separate Go `Entries` rendered in frontend)

### Pros
- No conceal/extmark work in GUI path.
- Frontend render model clean (`entries[]`), easy custom UI.
- Easy enrich later (git state, badges, virtual rows).
- Backend can normalize/path-resolve once.

### Cons
- Two truths: neovim buffer text **and** Go entries.
- Must sync every mutation path (`o/O/p/P/cc/dd/u/<C-r>`).
- Identity complexity for non-parseable lines.
- Bugs show as “buffer changed but UI stale”.

---

## Alternative directions

### Option A — Keep current architecture, switch to incremental line splice (recommended now)
- Keep `Entries` for frontend.
- Apply `nvim_buf_lines_event` diff directly.
- No full snapshot reads per keystroke.
- Still need internal pending IDs for keyed rendering stability.

Best for: minimal refactor, fast delivery.

### Option B — Current pane becomes row/buffer-first model
- Current pane source of truth = raw lines (+ cursor row).
- Parse id only when needed for file ops.
- Selection by row, not id.
- Removes “draft id” concept almost fully.

Best for: behavior closest to mini.files semantics.
Cost: bigger API/model refactor (`selectedEntryId` assumptions).

### Option C — Render current pane from grid directly
- Treat file explorer current pane like normal nvim content.
- Keep semantic model only for parent/preview ops.

Best for: strict single truth.
Cost: custom cursor/icon/entry UI harder, parse from tokens noisy.

---

## Replies to inline comments

> unclear to me why we need this and cant just return a DirectoryEntry

Agree. No dedicated helper struct required.

Use parser shape:
- `parseLine(line string) (entry DirectoryEntry, ok bool)`

Where:
- `ok=true` => parseable `<id>/text`
- `ok=false` => caller builds pending entry

So parser stays tiny.

---

> this seems crazy to me. why would i read the full buffer every time?
> `nvim_buf_lines_event` already tells me exactly which lines have changed.

Agree. v2 drops full-buffer rebuild.

Use incremental splice from event payload:
- replace `[firstline:lastline)` with `linedata`
- build only replacement segment
- preserve untouched prefix/suffix entries

Optional safety valve later: periodic full resync only on detected desync/debug mode.

---

> this is actually i a good point. i didnt think about this problem.
> when you type o then insert another character that will show up as a changed line in the event
> then you still cant parse an id, but you dont want to create a new entry because you already had created a draft id
> question is: how exactly does this work? how do you match the new changed line to the existing pending entry?

Mechanism in incremental splice:

1. Event gives replaced old range `[first:last)`.
2. For each new line `linedata[i]`:
   - if parseable id: use parsed id
   - else if `i < len(oldReplaced)` and `oldReplaced[i]` was pending: reuse `oldReplaced[i].ID`
   - else allocate new `nextID()`

Why this works:
- Typing on same pending row usually arrives as replace-1-with-1.
- Same slot in replaced range -> same pending ID reused.
- Insert/delete naturally shift surrounding rows via splice.

---

## Revised Decisions Locked

1. Keep raw buffer text unchanged.
2. Non-parseable line => pending entry.
3. Pending IDs internal only, allocated via `nextID()`.
4. Incremental `nvim_buf_lines_event` splice, not full-buffer reads.
5. Pending icon empty (`""`, fallback `" "` if needed).
6. Pending path `current.Path + "/" + text` (live recompute for changed lines).
7. Selected pending row sets `current.SelectedEntryId = pendingID`.
8. Pending selection clears preview.
9. Handle events only when explorer active + buf == current buffer.
10. Undo/redo handled through same line-event splice path.

---

## Incremental Update Plan

Given:
- `firstline`, `lastline`, `linedata[]` from `nvim_buf_lines_event`
- existing `e.current.Entries`

Process:
1. Validate/gate event buffer.
2. Clamp indices to current slice bounds.
3. Capture `oldReplaced := entries[first:last]`.
> not sure about this oldReplaced business. we'll see during implementation, but i feel like i might do it differently
4. Build `newSegment` from `linedata` using parse + pending-ID reuse rule.
5. Splice:
   - `entries = append(entries[:first], newSegment..., entries[last:]...)`
6. Update `e.current.Entries`, mark dirty.
7. If cursor row now points to pending line, keep selected pending ID.

Notes:
- Handle `linedata=[]` for deletions.
- Handle multi-line paste/replace.
- `more` batches can be applied sequentially; final state converges.

---

## Acceptance Criteria (v2)

- [x] `o` creates blank row; row renders immediately. @done(04/24/26 17:19)
- [x] Typing on pending row keeps same pending ID (no flicker). @done(04/24/26 17:19)
- [x] `O/p/P/cc` produce pending rows correctly. @done(04/24/26 17:19)
- [x] `dd` removes row immediately. @done(04/24/26 17:19)
- [x] `u/<C-r>` reflect immediately. @done(04/24/26 17:20)
- [x] Cursor move onto pending row never errors. @done(04/24/26 17:20)
- [x] Preview clears on pending selection. @done(04/24/26 17:20)
- [x] Only current pane buffer mutates explorer state. @done(04/24/26 17:20)

---

## Suggestion: likely better long-term model

If you want mini.files-like behavior with less sync burden:
- move current pane to **row-based state** (`selectedRow`, `rawLines[]`)
- treat parseable IDs as optional metadata
- keep parent pane/path ops semantic

Then “pending entry” stops being special concept. It becomes “line without id metadata”.
