# File Explorer Sync Future Notes

Not implementation plan for current slice.
Memory dump for later work so decisions/context do not get lost.

---

## Current agreed behavior

### Broken id semantics

If buffer line keeps parseable id, line can still be associated with previously known entry.

If buffer line loses parseable id, association is lost.
This is acceptable.

Examples:

- `Shift-C` rename preserves id prefix -> still recognized as rename-like edit
- `cc` destroys id -> line becomes pending/new-like row

### Empty line semantics

Empty line should remain visible if still present in buffer.
No layout jitter from replacing line content with empty text.

### Deletion semantics

If line is truly deleted from buffer (`dd`), row should disappear from frontend.
This is not considered unwanted jitter.
Reason: buffer structure changed, not only content.

### Sorting semantics

Do not resort while editing.
Buffer order is authoritative during edit session.
Resort only during later sync/reload flow if desired.

---

## Pending entry model

Need future model for rows that no longer map cleanly to original entry.
Examples:
- broken-id rows
- rows created with `o` / `O`
- pasted rows with no valid explorer id
- duplicate visible names

Possible future fields:
- `OriginalText`
- `OriginalPath`
- `DraftText`
- `DraftPath`
- `ParsedBufferID`
- `HasParsedBufferID`
- `State` (`known`, `pending-create`, `pending-rename`, `detached`, `duplicate`, ...)

No decision yet. Only note.

---

## Clipboard / paste edge cases

User called out important distinction:

### `o` / `O`
Always new pending entry.

### `p` / `P`
Depends on pasted content.

Examples:

#### Duplicate explorer line pasted unchanged
- `yy`
- `p`
- same file line appears twice
- future sync cannot create exact duplicate filesystem entry
- likely one original + one pending copy intent

#### Duplicate explorer line pasted then renamed
- `yy`
- `p`
- edit second line from `foo.txt` to `foo1.txt`
- future sync should produce two files with same content semantics if system supports copy-on-sync model

Need explicit later rules for:
- when pasted line represents copy
- when it represents noop duplicate
- how content source is inferred
- how hidden id duplication participates

---

## Preview semantics after identity loss

Current live-sync slice does not solve preview behavior fully.
Need later policy.

Possible policies:

1. **Strict association**
   - parseable id required
   - if id breaks, preview clears or switches to generic pending state

2. **Sticky preview**
   - preserve old preview until cursor moves
   - simpler visually, but semantically stale

3. **Draft-aware preview**
   - if detached row text matches existing nearby/original file, keep preview heuristically
   - likely too magic

For now, do not lock this.

---

## Parent/current selection consistency

User dislikes parent pane behaving fundamentally differently from current pane.
Need keep this in mind.

Even if implementation temporarily fail-softs current pane when id breaks, longer-term model should avoid surprising mismatch between panes.
Possible direction:
- every rendered row always has internal stable render id
- parsed hidden id becomes only one piece of metadata, not entire identity model

No current work required.

---

## Stable render identity concern

Frontend list rendering needs stable keys.
Hidden parsed id alone may be insufficient because:
- rows can lose id
- rows can be duplicated
- rows can be pasted or cleared

Future model likely needs:
- internal render id stable across rebuilds
- parsed buffer id stored separately

This is important even if not user-visible.
Without it, UI may churn during insert-mode edits.

---

## AttachBuffer hardening ideas

Later, after basic feature works:
- handle `on_reload`
- handle detach cleanup on explorer close
- possibly ignore pure `changedtick` events if text snapshot unchanged
- possibly debounce projector emission if needed, but probably not necessary given small buffers

---

## Sync-phase topics to revisit later

### Need diff model between:
- original directory snapshot
- current edited buffer snapshot

### Need filesystem operation categories:
- rename
- create
- delete
- copy
- move between directories
- duplicate/noop collapse

### Need conflict handling:
- duplicate filenames
- rename target already exists
- file deleted on disk externally
- case-only rename on case-insensitive fs
- dir/file type mismatches

### Need user feedback:
- confirm prompts
- error surfacing
- partial success reporting

### Need ordering decision:
- sync preserves buffer order?
- reload from filesystem re-sorts?
- pending rows merge where?

---

## Nice-to-have tests later

- parse valid `id/text`
- parse broken lines
- preserve empty lines
- rebuild after `cc`
- rebuild after `dd`
- rebuild after `o` / `O`
- rebuild after `p` / `P`
- selection update on broken-id line
- duplicate pasted rows

---

## Summary

Current implementation slice should stay narrow:
- mirror buffer text into `current.Entries`
- tolerate destroyed ids
- avoid overdesigning sync now

But future sync work must remember:
- broken-id rows are intentional
- duplicate rows are possible
- pasted explorer lines imply copy-like semantics later
- render identity and semantic identity are not same thing
