package fileexplorer

import (
	"fmt"
	"os"
	path "path/filepath"
	"testing"
)

type fakeNvim struct {
	linesByBuf map[int][][]byte
	calls      []string
}

type fakeEmitter struct {
	events []string
}

func (f *fakeEmitter) Emit(name string, payload any) {
	f.events = append(f.events, name)
}

func (f *fakeNvim) CreateBuffer(listed, scratch bool) (int, error) { return 0, nil }
func (f *fakeNvim) SetBufferLines(buf int, start, end int, strict bool, lines [][]byte) error {
	f.calls = append(f.calls, fmt.Sprintf("set:%d", buf))
	if f.linesByBuf == nil {
		f.linesByBuf = map[int][][]byte{}
	}
	copied := make([][]byte, len(lines))
	for i := range lines {
		copied[i] = append([]byte(nil), lines[i]...)
	}
	f.linesByBuf[buf] = copied
	return nil
}
func (f *fakeNvim) GetBufferLines(buf int, start, end int, strict bool) ([][]byte, error) {
	if lines, ok := f.linesByBuf[buf]; ok {
		return lines, nil
	}
	return [][]byte{}, nil
}
func (f *fakeNvim) Command(cmd string) error                               { return nil }
func (f *fakeNvim) OpenSplitRight(winId *int, bufNr int) error             { return nil }
func (f *fakeNvim) CurrentWindow() (int, error)                            { return 0, nil }
func (f *fakeNvim) SetWindowCursor(winId, row, col int) error              { return nil }
func (f *fakeNvim) SetCurrentWindow(winId int) error                       { return nil }
func (f *fakeNvim) SetBufferToWindow(winId int, bufNr int) error           { return nil }
func (f *fakeNvim) GetCurrentFilepath() (string, error)                    { return "", nil }
func (f *fakeNvim) SetWindowOption(winId int, key string, value any) error { return nil }
func (f *fakeNvim) CreateBufferKeymap(bufNr int, mode, lhs string, rhs func(channelID int) string) error {
	return nil
}
func (f *fakeNvim) CreateBufferAutocmd(winId, bufNr int, luaCallback string) error { return nil }
func (f *fakeNvim) AttachBuffer(bufNr int, sendBuffer bool, opts map[string]any) (bool, error) {
	f.calls = append(f.calls, fmt.Sprintf("attach:%d", bufNr))
	return true, nil
}
func (f *fakeNvim) DetachBuffer(bufNr int) (bool, error) {
	f.calls = append(f.calls, fmt.Sprintf("detach:%d", bufNr))
	return true, nil
}
func (f *fakeNvim) RegisterHandler(event string, handler func(data ...any)) {}

func TestBuildSyncActionPlan_MoveAcrossDirs(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		11: {},
		12: {[]byte("10/file.txt")},
	}}
	e := NewFileExplorer(nvim)
	e.sourceByID[10] = DirectoryEntry{ID: 10, Path: "/project/child/file.txt", Text: "file.txt", IsDir: false}
	e.dirtyByBuf[11] = &DirDraft{
		BufNr:   11,
		DirPath: "/project/child",
		OriginalEntries: map[uint64]DirectoryEntry{
			10: {ID: 10, Path: "/project/child/file.txt", Text: "file.txt", IsDir: false},
		},
	}
	e.dirtyByBuf[12] = &DirDraft{
		BufNr:           12,
		DirPath:         "/project",
		OriginalEntries: map[uint64]DirectoryEntry{},
	}

	actions, err := e.BuildSyncActionPlan()
	if err != nil {
		t.Fatalf("BuildSyncActionPlan failed: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d: %#v", len(actions), actions)
	}
	a := actions[0]
	if a.Kind != "move" || a.From != "/project/child/file.txt" || a.To != "/project/file.txt" {
		t.Fatalf("unexpected action: %#v", a)
	}
}

func TestBuildSyncActionPlan_CopyAcrossDirs(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		22: {[]byte("10/file.txt")},
	}}
	e := NewFileExplorer(nvim)
	e.sourceByID[10] = DirectoryEntry{ID: 10, Path: "/project/child/file.txt", Text: "file.txt", IsDir: false}
	e.dirtyByBuf[22] = &DirDraft{
		BufNr:           22,
		DirPath:         "/project",
		OriginalEntries: map[uint64]DirectoryEntry{},
	}

	actions, err := e.BuildSyncActionPlan()
	if err != nil {
		t.Fatalf("BuildSyncActionPlan failed: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d: %#v", len(actions), actions)
	}
	a := actions[0]
	if a.Kind != "copy" || a.From != "/project/child/file.txt" || a.To != "/project/file.txt" {
		t.Fatalf("unexpected action: %#v", a)
	}
}

func TestBuildSyncActionPlan_RenameInSameDir(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		33: {[]byte("10/renamed.txt")},
	}}
	e := NewFileExplorer(nvim)
	e.sourceByID[10] = DirectoryEntry{ID: 10, Path: "/project/file.txt", Text: "file.txt", IsDir: false}
	e.dirtyByBuf[33] = &DirDraft{
		BufNr:   33,
		DirPath: "/project",
		OriginalEntries: map[uint64]DirectoryEntry{
			10: {ID: 10, Path: "/project/file.txt", Text: "file.txt", IsDir: false},
		},
	}

	actions, err := e.BuildSyncActionPlan()
	if err != nil {
		t.Fatalf("BuildSyncActionPlan failed: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d: %#v", len(actions), actions)
	}
	a := actions[0]
	if a.Kind != "rename" || a.From != "/project/file.txt" || a.To != "/project/renamed.txt" {
		t.Fatalf("unexpected action: %#v", a)
	}
}

func TestApplySyncActions_CreateNestedFileCreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	e := NewFileExplorer(&fakeNvim{})
	target := path.Join(tmpDir, "foo", "foo.txt")

	err := e.applySyncActions([]SyncAction{{Kind: "create", To: target, IsDir: false}})
	if err != nil {
		t.Fatalf("applySyncActions failed: %v", err)
	}

	if _, err := os.Stat(path.Join(tmpDir, "foo")); err != nil {
		t.Fatalf("parent dir not created: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("target file not created: %v", err)
	}
}

func TestApplySyncActions_FailFast(t *testing.T) {
	tmpDir := t.TempDir()
	e := NewFileExplorer(&fakeNvim{})
	existing := path.Join(tmpDir, "a.txt")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatalf("write existing failed: %v", err)
	}
	second := path.Join(tmpDir, "b.txt")

	err := e.applySyncActions([]SyncAction{
		{Kind: "create", To: existing, IsDir: false},
		{Kind: "create", To: second, IsDir: false},
	})
	if err == nil {
		t.Fatalf("expected fail-fast error")
	}
	if _, err := os.Stat(second); !os.IsNotExist(err) {
		t.Fatalf("second action should not run, stat err=%v", err)
	}
}

func TestSync_CreateNestedFileFromDraftLine(t *testing.T) {
	tmpDir := t.TempDir()
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		88: {[]byte("foo/foo.txt")},
	}}
	e := NewFileExplorer(nvim)
	dir := &Directory{BufNr: 88, Path: tmpDir, Entries: []DirectoryEntry{}}
	e.directoriesByPath[tmpDir] = dir
	e.directoriesByBuf[88] = dir
	e.parent = dir
	e.current = dir
	e.parentWinID = 1
	e.currentWinID = 2
	e.previewWinID = 3
	e.dirtyByBuf[88] = &DirDraft{BufNr: 88, DirPath: tmpDir, OriginalEntries: map[uint64]DirectoryEntry{}}

	if err := e.Sync(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if _, err := os.Stat(path.Join(tmpDir, "foo", "foo.txt")); err != nil {
		t.Fatalf("nested file not created: %v", err)
	}
	if len(e.dirtyByBuf) != 0 {
		t.Fatalf("dirty drafts not cleared")
	}
	expectedCalls := []string{"detach:88", "set:88", "attach:88"}
	if len(nvim.calls) != len(expectedCalls) {
		t.Fatalf("expected nvim calls %#v, got %#v", expectedCalls, nvim.calls)
	}
	for i, expected := range expectedCalls {
		if nvim.calls[i] != expected {
			t.Fatalf("expected nvim calls %#v, got %#v", expectedCalls, nvim.calls)
		}
	}
}

func TestUpdateEntries_SameRowIDReplacementKeepsExistingEntryShape(t *testing.T) {
	nvim := &fakeNvim{}
	e := NewFileExplorer(nvim)
	dir := &Directory{BufNr: 7, Path: "/project", Entries: []DirectoryEntry{
		{ID: 10, Icon: DIR_ICON, Text: "old", Path: "/project/old", IsDir: true},
	}}
	e.directoriesByBuf[7] = dir

	if err := e.UpdateEntries(7, 0, 1, []string{"10/new"}); err != nil {
		t.Fatalf("UpdateEntries failed: %v", err)
	}

	if len(dir.Entries) != 1 {
		t.Fatalf("expected replacement row kept, got %d entries", len(dir.Entries))
	}
	entry := dir.Entries[0]
	if entry.ID != 10 || entry.Text != "new" || entry.Path != "/project/new" {
		t.Fatalf("unexpected replacement entry: %#v", entry)
	}
	if entry.Icon != DIR_ICON || !entry.IsDir || entry.IsDraft {
		t.Fatalf("replacement corrupted entry shape: %#v", entry)
	}
}

func TestSync_SelectiveRefreshSkipsUntouchedCachedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	touched := path.Join(tmpDir, "touched")
	untouched := path.Join(tmpDir, "untouched")
	if err := os.MkdirAll(touched, 0o755); err != nil {
		t.Fatalf("mkdir touched failed: %v", err)
	}
	if err := os.MkdirAll(untouched, 0o755); err != nil {
		t.Fatalf("mkdir untouched failed: %v", err)
	}
	if err := os.WriteFile(path.Join(untouched, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write untouched file failed: %v", err)
	}

	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		1: {[]byte("foo.txt")},
		2: {[]byte("20/existing.txt")},
	}}
	e := NewFileExplorer(nvim)
	touchedDir := &Directory{BufNr: 1, Path: touched, Entries: []DirectoryEntry{}}
	untouchedDir := &Directory{BufNr: 2, Path: untouched, Entries: []DirectoryEntry{
		{ID: 20, Icon: FILE_ICON, Text: "existing.txt", Path: path.Join(untouched, "existing.txt")},
	}}
	e.directoriesByPath[touched] = touchedDir
	e.directoriesByPath[untouched] = untouchedDir
	e.directoriesByBuf[1] = touchedDir
	e.directoriesByBuf[2] = untouchedDir
	e.parent = touchedDir
	e.current = touchedDir
	e.parentWinID = 1
	e.currentWinID = 2
	e.previewWinID = 3
	e.dirtyByBuf[1] = &DirDraft{BufNr: 1, DirPath: touched, OriginalEntries: map[uint64]DirectoryEntry{}}

	if err := e.Sync(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if _, err := os.Stat(path.Join(touched, "foo.txt")); err != nil {
		t.Fatalf("created file missing: %v", err)
	}
	for _, call := range nvim.calls {
		if call == "set:2" {
			t.Fatalf("untouched cached directory was rewritten, calls=%#v", nvim.calls)
		}
	}
}

func TestSetBufferLinesFromEntries_SkipsNoOpWrite(t *testing.T) {
	nvim := &fakeNvim{linesByBuf: map[int][][]byte{
		7: {[]byte("10/file.txt")},
	}}
	e := NewFileExplorer(nvim)

	err := e.setBufferLinesFromEntries(7, []DirectoryEntry{{ID: 10, Text: "file.txt"}})
	if err != nil {
		t.Fatalf("setBufferLinesFromEntries failed: %v", err)
	}
	if len(nvim.calls) != 0 {
		t.Fatalf("expected no SetBufferLines call, got %#v", nvim.calls)
	}
}

func TestMapDirectoryEntries_ReusesIDForStablePath(t *testing.T) {
	tmpDir := t.TempDir()
	stablePath := path.Join(tmpDir, "stable.txt")
	if err := os.WriteFile(stablePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write stable file failed: %v", err)
	}

	e := NewFileExplorer(&fakeNvim{})
	first, err := e.mapDirectoryEntries(tmpDir)
	if err != nil {
		t.Fatalf("first mapDirectoryEntries failed: %v", err)
	}
	if err := os.WriteFile(path.Join(tmpDir, "other.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write other file failed: %v", err)
	}
	second, err := e.mapDirectoryEntries(tmpDir)
	if err != nil {
		t.Fatalf("second mapDirectoryEntries failed: %v", err)
	}

	firstID := idForTestPath(t, first, stablePath)
	secondID := idForTestPath(t, second, stablePath)
	if firstID != secondID {
		t.Fatalf("expected stable ID for unchanged path, first=%d second=%d", firstID, secondID)
	}
}

func idForTestPath(t *testing.T, entries []DirectoryEntry, target string) uint64 {
	t.Helper()
	for _, entry := range entries {
		if entry.Path == target {
			return entry.ID
		}
	}
	t.Fatalf("path %s not found in %#v", target, entries)
	return 0
}

func TestRequestClose_ShowsPromptWhenDirty(t *testing.T) {
	emitter := &fakeEmitter{}
	e := NewFileExplorer(&fakeNvim{})
	e.SetEmitter(emitter)
	e.active = true
	e.dirtyByBuf[1] = &DirDraft{BufNr: 1, DirPath: "/project", OriginalEntries: map[uint64]DirectoryEntry{}}

	e.RequestClose()

	if !e.pendingClosePrompt {
		t.Fatalf("expected pending close prompt")
	}
	if len(emitter.events) == 0 || emitter.events[len(emitter.events)-1] != "file-explorer-confirm-prompt-show" {
		t.Fatalf("expected confirm prompt show event, got %#v", emitter.events)
	}
}

func TestHandleConfirmChoice_DiscardClosesExplorer(t *testing.T) {
	emitter := &fakeEmitter{}
	e := NewFileExplorer(&fakeNvim{})
	e.SetEmitter(emitter)
	e.active = true
	e.dirtyByBuf[1] = &DirDraft{BufNr: 1, DirPath: "/project", OriginalEntries: map[uint64]DirectoryEntry{}}

	e.RequestClose()
	e.HandleConfirmChoice(2)

	if e.active {
		t.Fatalf("expected explorer closed")
	}
}

func TestHandleConfirmChoice_CancelKeepsExplorerOpen(t *testing.T) {
	emitter := &fakeEmitter{}
	e := NewFileExplorer(&fakeNvim{})
	e.SetEmitter(emitter)
	e.active = true
	e.dirtyByBuf[1] = &DirDraft{BufNr: 1, DirPath: "/project", OriginalEntries: map[uint64]DirectoryEntry{}}

	e.RequestClose()
	e.HandleConfirmChoice(3)

	if !e.active {
		t.Fatalf("expected explorer to stay open on cancel")
	}
	if e.pendingClosePrompt {
		t.Fatalf("expected prompt hidden on cancel")
	}
}
