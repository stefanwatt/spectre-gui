package fileexplorer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPreviewLinesLimitsRows(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(file, []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, binary, _, err := readPreviewLines(file, 2)
	if err != nil {
		t.Fatalf("readPreviewLines failed: %v", err)
	}
	if binary {
		t.Fatal("expected text file")
	}
	if len(lines) != 2 || lines[0] != "one" || lines[1] != "two" {
		t.Fatalf("unexpected lines: %#v", lines)
	}
}

func TestReadPreviewLinesDetectsBinary(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sample.bin")
	if err := os.WriteFile(file, []byte{'a', 0, 'b'}, 0o644); err != nil {
		t.Fatal(err)
	}
	_, binary, _, err := readPreviewLines(file, 10)
	if err != nil {
		t.Fatalf("readPreviewLines failed: %v", err)
	}
	if !binary {
		t.Fatal("expected binary file")
	}
}

func TestIsPreviewImage(t *testing.T) {
	if !isPreviewImage("photo.WEBP") {
		t.Fatal("expected webp image")
	}
	if isPreviewImage("notes.txt") {
		t.Fatal("did not expect text file image")
	}
}

func TestRefreshDirectoryPreviewUsesPureData(t *testing.T) {
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("mkdir child failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(childDir, "nested.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write nested file failed: %v", err)
	}

	nvim := &fakeNvim{nextBuf: 200}
	e := NewFileExplorer(nvim)
	e.preview = Directory{BufNr: 99}
	e.previewWinID = 3
	e.currentWinID = 2

	entry := DirectoryEntry{ID: 1, Text: "child", Path: childDir, IsDir: true}
	if err := e.refreshDirectoryPreview(entry); err != nil {
		t.Fatalf("refreshDirectoryPreview failed: %v", err)
	}
	if e.previewKind != PreviewKindDirectory {
		t.Fatalf("expected directory preview kind, got %q", e.previewKind)
	}
	if e.preview.BufNr != 99 {
		t.Fatalf("directory preview must preserve neutral preview buffer, got %d", e.preview.BufNr)
	}
	if e.preview.Path != childDir || e.previewTitle != "child" || e.previewPath != childDir {
		t.Fatalf("unexpected preview metadata: preview=%#v title=%q path=%q", e.preview, e.previewTitle, e.previewPath)
	}
	if len(e.preview.Entries) != 1 || e.preview.Entries[0].Text != "nested.txt" {
		t.Fatalf("unexpected preview entries: %#v", e.preview.Entries)
	}
	if _, ok := e.directoriesByPath[childDir]; ok {
		t.Fatalf("directory preview registered canonical directory")
	}
	if len(e.directoriesByBuf) != 0 {
		t.Fatalf("directory preview registered buffer mapping: %#v", e.directoriesByBuf)
	}
	if nvim.createBufferCallCount != 0 {
		t.Fatalf("directory preview created nvim buffers")
	}
	if len(nvim.bufferWindowCalls) != 0 {
		t.Fatalf("directory preview swapped nvim window buffers: %#v", nvim.bufferWindowCalls)
	}
}

func TestRefreshDirectoryPreviewDoesNotDeleteDisplayedTextBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("mkdir child failed: %v", err)
	}

	nvim := &fakeNvim{}
	e := NewFileExplorer(nvim)
	e.preview = Directory{BufNr: 99}
	e.previewWinID = 3
	e.currentWinID = 2
	e.previewTextBufNr = 42
	e.previewTextBufPath = filepath.Join(tmpDir, "old.txt")

	entry := DirectoryEntry{ID: 1, Text: "child", Path: childDir, IsDir: true}
	if err := e.refreshDirectoryPreview(entry); err != nil {
		t.Fatalf("refreshDirectoryPreview failed: %v", err)
	}
	if nvim.deleteBufferCallCount != 0 {
		t.Fatalf("directory preview deleted displayed text buffer")
	}
	if e.previewTextBufNr != 42 {
		t.Fatalf("expected text preview buffer kept alive, got %d", e.previewTextBufNr)
	}
}

func TestRefreshPreviewForEntryDoesNotLeaveOldDirectoryOnTextError(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "notes.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write text file failed: %v", err)
	}

	e := NewFileExplorer(&fakeNvim{createBufferErr: os.ErrPermission})
	e.active = true
	e.previewGeneration = 1
	e.previewWinID = 3
	e.currentWinID = 2
	e.preview = Directory{BufNr: 99, Entries: []DirectoryEntry{{ID: 10, Text: "stale", IsDir: true}}}
	e.previewKind = PreviewKindDirectory
	e.previewPath = filepath.Join(tmpDir, "old-dir")
	e.previewTitle = "old-dir"

	entry := DirectoryEntry{ID: 2, Text: "notes.txt", Path: file}
	if err := e.refreshPreviewForEntry(entry, 1); err == nil {
		t.Fatalf("expected text preview setup error")
	}
	if e.previewKind != PreviewKindTextFile {
		t.Fatalf("expected text preview kind after failure, got %q", e.previewKind)
	}
	if e.previewPath != file || e.previewTitle != "notes.txt" {
		t.Fatalf("expected new file preview metadata, path=%q title=%q", e.previewPath, e.previewTitle)
	}
	if len(e.preview.Entries) != 0 {
		t.Fatalf("expected stale directory entries cleared, got %#v", e.preview.Entries)
	}
	if !e.Dirty {
		t.Fatalf("expected failed preview transition to mark explorer dirty")
	}
}

func TestRefreshPreviewForEntryReusesTextBufferAfterPathChange(t *testing.T) {
	tmpDir := t.TempDir()
	oldFile := filepath.Join(tmpDir, "old.txt")
	newFile := filepath.Join(tmpDir, "new.txt")
	if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
		t.Fatalf("write old file failed: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new"), 0o644); err != nil {
		t.Fatalf("write new file failed: %v", err)
	}

	nvim := &fakeNvim{nextBuf: 20}
	e := NewFileExplorer(nvim)
	e.active = true
	e.previewGeneration = 1
	e.previewWinID = 3
	e.currentWinID = 2
	e.preview = Directory{BufNr: 99}
	e.previewKind = PreviewKindTextFile
	e.previewPath = oldFile
	e.previewTitle = "old.txt"
	e.previewTextBufNr = 10
	e.previewTextBufPath = oldFile

	entry := DirectoryEntry{ID: 2, Text: "new.txt", Path: newFile}
	if err := e.refreshPreviewForEntry(entry, 1); err != nil {
		t.Fatalf("refreshPreviewForEntry failed: %v", err)
	}
	if e.previewTextBufNr != 10 || e.previewTextBufPath != newFile {
		t.Fatalf("expected reused text buffer for new path, buf=%d path=%q", e.previewTextBufNr, e.previewTextBufPath)
	}
	if got := string(nvim.linesByBuf[10][0]); got != "new" {
		t.Fatalf("expected new file content in reused preview buffer, got %q", got)
	}
	if nvim.createBufferCallCount != 0 {
		t.Fatalf("expected no new preview buffer, got %d create calls", nvim.createBufferCallCount)
	}
	if len(nvim.bufferWindowCalls) == 0 || nvim.bufferWindowCalls[len(nvim.bufferWindowCalls)-1].bufNr != 10 {
		t.Fatalf("expected preview window to use reused buffer, calls=%#v", nvim.bufferWindowCalls)
	}
}
