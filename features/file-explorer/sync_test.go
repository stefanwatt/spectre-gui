package fileexplorer

import (
	"os"
	path "path/filepath"
	"testing"
)

type fakeNvim struct {
	linesByBuf map[int][][]byte
}

type fakeEmitter struct {
	events []string
}

func (f *fakeEmitter) Emit(name string, payload any) {
	f.events = append(f.events, name)
}

func (f *fakeNvim) CreateBuffer(listed, scratch bool) (int, error) { return 0, nil }
func (f *fakeNvim) SetBufferLines(buf int, start, end int, strict bool, lines [][]byte) error {
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
