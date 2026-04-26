package fileexplorer

import "testing"

func TestBufferLineCursorMapping(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		start   int
		visible string
	}{
		{name: "prefixed", line: "12/foo", start: 3, visible: "foo"},
		{name: "bare", line: "foo", start: 0, visible: "foo"},
		{name: "numeric filename", line: "123", start: 0, visible: "123"},
		{name: "path draft", line: "foo/bar", start: 0, visible: "foo/bar"},
		{name: "leading space", line: " 12/foo", start: 0, visible: " 12/foo"},
		{name: "space after prefix", line: "12/ foo", start: 3, visible: " foo"},
		{name: "empty visible", line: "12/", start: 3, visible: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bufferLineVisibleStartCol(tt.line); got != tt.start {
				t.Fatalf("expected start %d, got %d", tt.start, got)
			}
			if got := bufferLineVisibleText(tt.line); got != tt.visible {
				t.Fatalf("expected visible %q, got %q", tt.visible, got)
			}
		})
	}
}

func TestBufferColToVisibleCol(t *testing.T) {
	tests := []struct {
		line string
		raw  int
		want int
	}{
		{line: "12/foo", raw: 0, want: 0},
		{line: "12/foo", raw: 3, want: 0},
		{line: "12/foo", raw: 4, want: 1},
		{line: "12/foo", raw: 99, want: 3},
		{line: "foo", raw: 2, want: 2},
		{line: "foo", raw: 99, want: 3},
	}

	for _, tt := range tests {
		if got := bufferColToVisibleCol(tt.line, tt.raw); got != tt.want {
			t.Fatalf("bufferColToVisibleCol(%q, %d) = %d, want %d", tt.line, tt.raw, got, tt.want)
		}
	}
}

func TestVisibleColToBufferCol(t *testing.T) {
	tests := []struct {
		line    string
		visible int
		want    int
	}{
		{line: "12/foo", visible: 0, want: 3},
		{line: "12/foo", visible: 1, want: 4},
		{line: "12/foo", visible: 99, want: 6},
		{line: "foo", visible: 2, want: 2},
		{line: "foo", visible: 99, want: 3},
	}

	for _, tt := range tests {
		if got := visibleColToBufferCol(tt.line, tt.visible); got != tt.want {
			t.Fatalf("visibleColToBufferCol(%q, %d) = %d, want %d", tt.line, tt.visible, got, tt.want)
		}
	}
}

func TestUpdateSelectedEntryStoresVisibleColForPrefixedLine(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		20: {[]byte("12/target.go")},
	}}
	e := NewFileExplorer(nvim)
	e.currentWinID = 102
	e.current = &Directory{BufNr: 20, Entries: []DirectoryEntry{{ID: 12, Text: "target.go"}}}

	if err := e.UpdateSelectedyEntry(0, 6); err != nil {
		t.Fatalf("UpdateSelectedyEntry failed: %v", err)
	}

	if e.current.CursorCol != 3 {
		t.Fatalf("expected visible cursor col 3, got %d", e.current.CursorCol)
	}
	if len(nvim.cursorCalls) != 0 {
		t.Fatalf("expected no correction, got %#v", nvim.cursorCalls)
	}
}

func TestUpdateSelectedEntryStoresVisibleColForPendingLine(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		20: {[]byte("pending.txt")},
	}}
	e := NewFileExplorer(nvim)
	e.currentWinID = 102
	e.current = &Directory{BufNr: 20, Entries: []DirectoryEntry{{ID: 99, Text: "pending.txt", IsDraft: true}}}

	if err := e.UpdateSelectedyEntry(0, 4); err != nil {
		t.Fatalf("UpdateSelectedyEntry failed: %v", err)
	}

	if e.current.CursorCol != 4 {
		t.Fatalf("expected visible cursor col 4, got %d", e.current.CursorCol)
	}
	if len(nvim.cursorCalls) != 0 {
		t.Fatalf("expected no correction, got %#v", nvim.cursorCalls)
	}
}

func TestUpdateSelectedEntryCorrectsCursorBeforeNameStart(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		20: {[]byte("12/target.go")},
	}}
	e := NewFileExplorer(nvim)
	e.currentWinID = 102
	e.current = &Directory{
		BufNr:           20,
		SelectedEntryId: 12,
		Entries:         []DirectoryEntry{{ID: 12, Text: "target.go"}},
	}

	if err := e.UpdateSelectedyEntry(0, 1); err != nil {
		t.Fatalf("UpdateSelectedyEntry failed: %v", err)
	}

	if e.current.CursorCol != 0 {
		t.Fatalf("expected visible cursor col 0, got %d", e.current.CursorCol)
	}
	expected := []cursorCall{{winID: 102, row: 1, col: 3}}
	if len(nvim.cursorCalls) != len(expected) || nvim.cursorCalls[0] != expected[0] {
		t.Fatalf("expected correction %#v, got %#v", expected, nvim.cursorCalls)
	}
}

func TestUpdateSelectedEntryPreservesVisibleColAcrossRowChange(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		20: {[]byte("9/first.txt"), []byte("123/second.txt")},
	}}
	e := NewFileExplorer(nvim)
	e.currentWinID = 102
	e.current = &Directory{
		BufNr:           20,
		SelectedEntryId: 9,
		CursorCol:       4,
		Entries: []DirectoryEntry{
			{ID: 9, Text: "first.txt"},
			{ID: 123, Text: "second.txt"},
		},
	}

	if err := e.UpdateSelectedyEntry(1, 4); err != nil {
		t.Fatalf("UpdateSelectedyEntry failed: %v", err)
	}

	if e.current.CursorCol != 4 {
		t.Fatalf("expected visible cursor col 4, got %d", e.current.CursorCol)
	}
	expected := []cursorCall{{winID: 102, row: 2, col: 8}}
	if len(nvim.cursorCalls) != len(expected) || nvim.cursorCalls[0] != expected[0] {
		t.Fatalf("expected correction %#v, got %#v", expected, nvim.cursorCalls)
	}
}

func TestUpdateSelectedEntryClampsPreservedVisibleColToShorterRow(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		20: {[]byte("9/long-name"), []byte("123/go")},
	}}
	e := NewFileExplorer(nvim)
	e.currentWinID = 102
	e.current = &Directory{
		BufNr:           20,
		SelectedEntryId: 9,
		CursorCol:       8,
		Entries: []DirectoryEntry{
			{ID: 9, Text: "long-name"},
			{ID: 123, Text: "go"},
		},
	}

	if err := e.UpdateSelectedyEntry(1, 4); err != nil {
		t.Fatalf("UpdateSelectedyEntry failed: %v", err)
	}

	if e.current.CursorCol != 2 {
		t.Fatalf("expected visible cursor col 2, got %d", e.current.CursorCol)
	}
	expected := []cursorCall{{winID: 102, row: 2, col: 6}}
	if len(nvim.cursorCalls) != len(expected) || nvim.cursorCalls[0] != expected[0] {
		t.Fatalf("expected correction %#v, got %#v", expected, nvim.cursorCalls)
	}
}

func TestSyncPaneCursorToSelectionMapsVisibleColToRawCol(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		20: {[]byte("12/target.go")},
	}}
	e := NewFileExplorer(nvim)
	dir := &Directory{
		BufNr:           20,
		SelectedEntryId: 12,
		CursorCol:       3,
		Entries:         []DirectoryEntry{{ID: 12, Text: "target.go"}},
	}

	if err := e.syncPaneCursorToSelection(102, dir); err != nil {
		t.Fatalf("syncPaneCursorToSelection failed: %v", err)
	}

	expected := []cursorCall{{winID: 102, row: 1, col: 6}}
	if len(nvim.cursorCalls) != len(expected) || nvim.cursorCalls[0] != expected[0] {
		t.Fatalf("expected cursor calls %#v, got %#v", expected, nvim.cursorCalls)
	}
}

func TestSyncPaneCursorToSelectionMapsNoPrefixLine(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		20: {[]byte("pending.txt")},
	}}
	e := NewFileExplorer(nvim)
	dir := &Directory{
		BufNr:           20,
		SelectedEntryId: 99,
		CursorCol:       2,
		Entries:         []DirectoryEntry{{ID: 99, Text: "pending.txt", IsDraft: true}},
	}

	if err := e.syncPaneCursorToSelection(102, dir); err != nil {
		t.Fatalf("syncPaneCursorToSelection failed: %v", err)
	}

	expected := []cursorCall{{winID: 102, row: 1, col: 2}}
	if len(nvim.cursorCalls) != len(expected) || nvim.cursorCalls[0] != expected[0] {
		t.Fatalf("expected cursor calls %#v, got %#v", expected, nvim.cursorCalls)
	}
}
