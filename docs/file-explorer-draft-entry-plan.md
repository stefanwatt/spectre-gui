# File Explorer Pending Entry Plan

## Goal

Show new row in frontend whenever current explorer buffer contains line without parseable `<id>/` prefix.

Target behavior:
- `o`, `O`, `p`, `P`, `cc`, any edit path works
- row with non-parseable id becomes **pending entry**
- pending row renders immediately (even empty)
- no filesystem sync yet

---

## Decisions Locked

1. **Draft IDs use `nextID()`**
   - Same counter as regular entries.
   - Reason: simple, collision-safe in-process, no extra ID system.

2. **Buffer text stays raw**
   - Do not inject draft id into neovim buffer line.
   - Pending id exists only in Go internal state.

3. **Pending detection rule**
   - Line parse fails for `<id>/...` → pending entry.
   - Creation source irrelevant (`o/O/p/P/normal edits`).

4. **Pending icon**
   - Empty icon (`""`) preferred.
   - Fallback single space if UI requires visible glyph.

5. **Pending path**
   - `Path = current.Path + "/" + text`.
   - Recompute live on each rebuild.

6. **Selection on pending row**
   - Set `current.SelectedEntryId = draftID`.

7. **Preview on pending select**
   - Clear preview.

8. **Duplicate id handling**
   - Ignore for this slice.

9. **Buf event guard**
   - Process `nvim_buf_lines_event` only when:
     - explorer active
     - event buf == `current.BufNr`

10. **Undo/redo**
   - Must reflect immediately via same buffer-event pipeline.

---

## Scope

In scope:
- current pane live rebuild from buffer snapshot
- pending entry projection to frontend
- safe cursor/selection behavior on pending lines
- preview clear policy for pending selection

Out of scope:
- filesystem apply/sync
- duplicate semantic reconciliation
- git/icon enrichment
- parent/preview pane editing

---

## Data Model Plan

## `DirectoryEntry`
Keep struct shape. Use existing fields:
- `ID`: real id or internal draft id
- `Text`: filename text (can be empty)
- `Path`: `current.Path + "/" + Text`
- `IsDir`: `false` for pending
- `Icon`: `""` (or `" "` fallback)

No new frontend type required for this slice.

## Internal parsing result (new helper model)
Introduce internal line parse result with explicit kind:
- `hasID bool`
- `id uint64` (valid when `hasID`)
- `text string`

Reason: current `parseEntryFromBufferLine` hard-fails. Pending lines need non-fatal parse path.

> unclear to me why we need this and cant just return a DirectoryEntry 
---

## Rebuild Strategy

Use full-buffer rebuild on each relevant `nvim_buf_lines_event`.

Flow:
1. Read full current buffer lines.
2. Parse each line:
   - parseable `<id>/text` → existing entry
   - non-parseable → pending entry
3. Build `e.current.Entries` in exact buffer order.
4. Mark explorer dirty.

Reason:
- Handles all edit shapes uniformly.
- Keeps impl small.
- Supports undo/redo naturally.

> this seems crazy to me. why would i read the full buffer every time?
`nvim_buf_lines_event` already tells me exactly which lines have changed.
---

## Draft ID Stability Plan

Need stable pending IDs across keystrokes. Unstable IDs cause keyed-list churn + selection flicker.

Strategy for this slice:
- Reuse previous pending ID when same row index was pending in previous snapshot.
- Allocate new `nextID()` when row has no reusable pending ID.

Notes:
- Good for typing in-place (`o` then insert text).
- May remap on complex row shifts. Accept for now.
- Can harden later with smarter row matching.

> this is actually i a good point. i didnt think about this problem. 
when you type o then insert another character that will show up as a changed line in the event
then you still cant parse an id, but you dont want to create a new entry because you already had created a draft id
question is: how exactly does this work? how do you match the new changed line to the existing pending entry?
---

## Cursor + Selection Plan

Current `UpdateSelectedyEntryCurrent` depends on strict id parse from buffer line.

Planned behavior:
- Resolve selected row against rebuilt `e.current.Entries` by row index.
- If row maps to pending entry, use pending draft ID.
- Update `CursorCol` always.
- Never error/panic when line lacks parseable id.

Reason: pending rows first-class for this slice.

---

## Preview Plan

When selected entry is pending:
- clear preview entries/content
- mark dirty

Reason: pending row has no trusted filesystem identity yet.

---

## Event Handling Plan

`nvim_buf_lines_event` handler:
- Validate payload shape.
- Extract buf id.
- Return early unless `GetFileExplorer().GetActive()` and `buf == current.BufNr`.
- Trigger current-pane rebuild.

Reason: avoid cross-buffer noise and accidental state corruption.

---

## Acceptance Criteria

1. `o` creates blank new row in current pane; frontend shows row immediately.
2. Typing on pending row updates text live.
3. `O`, `p`, `P`, `cc` produce pending rows same way.
4. `dd` removes row immediately from frontend.
5. Undo/redo updates rows immediately.
6. Cursor on pending row does not error.
7. Selected pending row uses draft id and stays highlighted.
8. Preview clears when pending row selected.
9. only buffer of pane "current" mutates explorer state.

---

## Manual Test Matrix

1. Open explorer, hit `o` in current pane.
   - Expect new blank row rendered.
2. Type `foo.md`.
   - Expect live text + path `current.Path/foo.md` internally.
3. Hit `Esc`, `u`, `<C-r>`.
   - Expect row/state mirror each step.
4. Paste multiple plain lines with `p`/`P`.
   - Expect pending rows rendered for non-parseable lines.
5. Move cursor across parseable and pending rows.
   - Expect no parse errors.
6. Trigger edits in non-current buffer.
   - Expect no explorer mutation.

---

## Risks

1. **Row-index pending ID reuse weak on complex shifts**
   - May reassign pending IDs after large insert/delete above.
   - Accepted for now.

2. **Duplicate parseable IDs from pasted lines**
   - Not handled now.

3. **Pending `Path` may be invalid/empty-text path**
   - Accepted until sync semantics added.

---

## Next Plan (Later)

- stronger pending ID reuse (content+neighbor matching)
- explicit pending metadata (`IsPending`, `OriginalID`)
- duplicate id normalization
- preview policy improvements for pending rows
- filesystem sync pipeline
