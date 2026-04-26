package projection

import (
	"os"
	"path/filepath"
	"testing"

	"nvim-gui/core/model"
	fileexplorer "nvim-gui/features/file-explorer"
	"nvim-gui/rendering"
)

type projectorFakeNvim struct {
	linesByBuf  map[int][][]byte
	nextBuf     int
	nextWin     int
	currentWin  int
	currentFile string
}

func (f *projectorFakeNvim) CreateBuffer(listed, scratch bool) (int, error) {
	if f.nextBuf == 0 {
		f.nextBuf = 10
	}
	bufNr := f.nextBuf
	f.nextBuf++
	return bufNr, nil
}

func (f *projectorFakeNvim) SetBufferLines(buf int, start, end int, strict bool, lines [][]byte) error {
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

func (f *projectorFakeNvim) GetBufferLines(buf int, start, end int, strict bool) ([][]byte, error) {
	lines := f.linesByBuf[buf]
	if start < 0 {
		start = 0
	}
	if end < 0 || end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) || start > end {
		return [][]byte{}, nil
	}
	return lines[start:end], nil
}

func (f *projectorFakeNvim) Command(cmd string) error { return nil }

func (f *projectorFakeNvim) OpenSplitRight(winId *int, bufNr int) error {
	if f.nextWin == 0 {
		f.nextWin = 100
	}
	*winId = f.nextWin
	f.nextWin++
	return nil
}

func (f *projectorFakeNvim) CurrentWindow() (int, error) {
	if f.currentWin == 0 {
		f.currentWin = 99
	}
	return f.currentWin, nil
}

func (f *projectorFakeNvim) SetWindowCursor(winId, row, col int) error    { return nil }
func (f *projectorFakeNvim) SetCurrentWindow(winId int) error             { return nil }
func (f *projectorFakeNvim) SetBufferToWindow(winId int, bufNr int) error { return nil }
func (f *projectorFakeNvim) GetCurrentFilepath() (string, error)          { return f.currentFile, nil }
func (f *projectorFakeNvim) SetWindowOption(winId int, key string, value any) error {
	return nil
}
func (f *projectorFakeNvim) SetBufferOption(bufNr int, key string, value any) error {
	return nil
}
func (f *projectorFakeNvim) DeleteBuffer(bufNr int, force bool) error                { return nil }
func (f *projectorFakeNvim) SetWindowSize(winId, cols, rows int) error               { return nil }
func (f *projectorFakeNvim) CreateBufferAutocmd(winId, bufNr int, lua string) error  { return nil }
func (f *projectorFakeNvim) RegisterHandler(event string, handler func(data ...any)) {}

func (f *projectorFakeNvim) ExecLua(script string, result any, args ...any) error {
	if out, ok := result.(*string); ok {
		*out = ""
	}
	return nil
}

func (f *projectorFakeNvim) CreateBufferKeymap(bufNr int, mode, lhs string, rhs func(channelID int) string) error {
	return nil
}

func (f *projectorFakeNvim) AttachBuffer(bufNr int, sendBuffer bool, opts map[string]any) (bool, error) {
	return true, nil
}

func (f *projectorFakeNvim) DetachBuffer(bufNr int) (bool, error) { return true, nil }

func TestFileExplorerTextPreviewFallsBackToStoredLinesWithoutGrid(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "notes.txt")
	if err := os.WriteFile(file, []byte("alpha\nbeta\n"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	fe := fileexplorer.NewFileExplorer(&projectorFakeNvim{currentFile: file})
	if err := fe.Open(nil); err != nil {
		t.Fatalf("open file explorer failed: %v", err)
	}

	projection := NewFileExplorerProjector(fe).Project(model.NewAppState())
	if len(projection.Events) != 1 {
		t.Fatalf("expected one file explorer event, got %#v", projection.Events)
	}
	payload, ok := projection.Events[0].Payload.(map[string]any)
	if !ok {
		t.Fatalf("unexpected payload type %T", projection.Events[0].Payload)
	}
	preview, ok := payload["preview"].(map[string]any)
	if !ok {
		t.Fatalf("missing preview payload: %#v", payload)
	}
	if preview["kind"] != "textFile" {
		t.Fatalf("expected text preview, got %#v", preview)
	}
	contentRows, ok := preview["content"].([]rendering.ContentRow)
	if !ok {
		t.Fatalf("unexpected content type %T", preview["content"])
	}
	if len(contentRows) != 2 {
		t.Fatalf("expected fallback rows, got %#v", contentRows)
	}
	if contentRows[0].Tokens[0].Text != "alpha" || contentRows[1].Tokens[0].Text != "beta" {
		t.Fatalf("unexpected fallback content: %#v", contentRows)
	}
}
