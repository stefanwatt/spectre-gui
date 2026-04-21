package fileexplorer

import (
	"fmt"
	"os"
	path "path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"nvim-gui/core/ports"
	"nvim-gui/utils"
)

type FileExplorer struct {
	active    bool
	Dirty     bool
	idCounter uint64
	nvim      ports.NvimClient
	parent    Directory
	current   Directory
	preview   Directory
}

type DirectoryEntry struct {
	ID        uint64 `json:"id"`
	Icon      string `json:"icon"`
	IconClass string `json:"iconClass"`
	Text      string `json:"text"`
	Path      string `json:"path"`
	IsDir     bool   `json:"isDir"`
}

// Directory holds parsed entries for one mini.files pane (parent/current/preview).
type Directory struct {
	WinID           int              `json:"winId"`
	BufNr           int              `json:"bufNr"`
	Entries         []DirectoryEntry `json:"entries"`
	SelectedEntryId uint64           `json:"selectedEntryId"`
	CursorCol       int              `json:"cursorCol"`
	Path            string
}

func NewFileExplorer(nvim ports.NvimClient) *FileExplorer {
	return &FileExplorer{
		nvim: nvim,
	}
}

func (e *FileExplorer) nextID() uint64 {
	return atomic.AddUint64(&e.idCounter, 1)
}

func (e *FileExplorer) GetParent() Directory {
	return e.parent
}

func (e *FileExplorer) GetCurrent() Directory {
	return e.current
}

func (e *FileExplorer) GetPreview() Directory {
	return e.preview
}

func (e *FileExplorer) GetActive() bool {
	return e.active
}

func (e *FileExplorer) GetCurrentBuf() int {
	return e.current.BufNr
}

func (e *FileExplorer) Open(_filepath *string) error {
	var err error
	var filepath string
	if _filepath == nil {
		filepath, err = e.nvim.GetCurrentFilepath()
	}
	if err != nil {
		return err
	}
	e.parent.BufNr, err = e.nvim.CreateBuffer(true, false)
	if err != nil {
		return err
	}
	e.current.BufNr, err = e.nvim.CreateBuffer(true, false)
	if err != nil {
		return err
	}
	e.preview.BufNr, err = e.nvim.CreateBuffer(true, false)
	if err != nil {
		return err
	}
	e.current.Path = path.Dir(filepath)
	e.parent.Path = path.Dir(e.current.Path)
	if e.parent.Path == "." {
		return fmt.Errorf("[FileExplorer] parent dir doesnt exist")
	}
	e.parent.Entries, err = e.mapDirectoryEntries(e.parent.Path)
	if err != nil {
		return err
	}
	e.current.Entries, err = e.mapDirectoryEntries(e.current.Path)
	if err != nil {
		return err
	}
	err = e.nvim.Command("tab new")
	if err != nil {
		return err
	}
	e.parent.WinID, err = e.nvim.CurrentWindow()
	if err != nil {
		return err
	}
	err = e.nvim.SetBufferToWindow(e.parent.WinID, e.parent.BufNr)
	if err != nil {
		return err
	}
	err = e.nvim.OpenSplitRight(&e.current.WinID, e.current.BufNr)
	if err != nil {
		return err
	}
	err = e.nvim.OpenSplitRight(&e.preview.WinID, e.preview.BufNr)
	if err != nil {
		return err
	}

	err = e.updateSelectedEntries(filepath)
	if err != nil {
		return err
	}
	e.preview.Entries = []DirectoryEntry{}

	err = e.syncPaneBuffers()
	if err != nil {
		return err
	}

	err = e.syncPaneCursorToSelection(e.parent.WinID, e.parent.BufNr, e.parent.SelectedEntryId)
	if err != nil {
		return err
	}

	err = e.syncPaneCursorToSelection(e.current.WinID, e.current.BufNr, e.current.SelectedEntryId)
	if err != nil {
		return err
	}

	e.active = true
	e.Dirty = true
	err = e.setupKeymaps()
	if err != nil {
		return err
	}
	e.nvim.SetCurrentWindow(e.current.WinID)
	//TODO: still getting panic 
	// panic: [UpdateCursor] no line at row=1
	// have to investigate
	// also have to change the buffer lines actually
	// i thought writing the entry id would be enough, but its not
	// if i only have the id then i could never edit
	// i suppose that also means i have to conceal the id after all, not sure tho
	// have to test how it is in mini.files and how conceal works
	// i just had the thought that conceal might protect me from clearing the id
	// with something like cc, but im not even sure that works in mini.files
	// and im not sure that conceal even prevents that
	// might have to do some cursor manipulation

	return err
}

func (e *FileExplorer) UpdateCursor(row, col int) error {
	if row < 0 || col < 0 {
		return fmt.Errorf("[UpdateCursor] row/col out of bounds")
	}

	lines, err := e.nvim.GetBufferLines(e.current.BufNr, row, row+1, false)
	if err != nil {
		return fmt.Errorf("[UpdateCursor] get buffer line failed: %w", err)
	}
	if len(lines) == 0 {
		return fmt.Errorf("[UpdateCursor] no line at row=%d", row)
	}

	line := strings.TrimSpace(string(lines[0]))
	if line == "" {
		return fmt.Errorf("[UpdateCursor] empty line at row=%d", row)
	}

	selectedID, err := strconv.ParseUint(line, 10, 64)
	if err != nil {
		return fmt.Errorf("[UpdateCursor] parse id from line '%s' failed: %w", line, err)
	}

	_, err = utils.Find(e.current.Entries, func(entry DirectoryEntry) bool {
		return entry.ID == selectedID
	})
	if err != nil {
		return fmt.Errorf("[UpdateCursor] id %d not in current entries", selectedID)
	}

	cursorChanged := e.current.CursorCol != col
	selectionChanged := e.current.SelectedEntryId != selectedID
	e.current.CursorCol = col
	e.current.SelectedEntryId = selectedID
	if cursorChanged || selectionChanged {
		e.Dirty = true
	}
	return nil
}

func (e *FileExplorer) updateSelectedEntries(filepath string) error {
	var err error
	var entry DirectoryEntry
	entry, err = utils.Find(e.parent.Entries, func(entry DirectoryEntry) bool {
		return strings.Contains(filepath, entry.Path)
	})
	if err != nil {
		return err
	}
	e.parent.SelectedEntryId = entry.ID

	entry, err = utils.Find(e.current.Entries, func(entry DirectoryEntry) bool {
		return filepath == entry.Path
	})
	e.current.SelectedEntryId = entry.ID
	return err
}

func (e *FileExplorer) setupKeymaps() error {
	err := e.nvim.CreateBufferKeymap(
		e.current.BufNr,
		"n",
		"q",
		func(channelID int) string {
			return fmt.Sprintf(":lua vim.rpcnotify(%d, 'FileExplorerClose', {})<CR>", channelID)
		},
	)
	if err != nil {
		return err
	}

	err = e.nvim.CreateBufferKeymap(
		e.current.BufNr,
		"n",
		"<Right>",
		func(channelID int) string {
			return fmt.Sprintf(":lua vim.rpcnotify(%d, 'FileExplorerGoIn', {})<CR>", channelID)
		},
	)

	if err != nil {
		return err
	}
	err = e.nvim.CreateBufferKeymap(
		e.current.BufNr,
		"n",
		"<Left>",
		func(channelID int) string {
			return fmt.Sprintf(":lua vim.rpcnotify(%d, 'FileExplorerGoOut', {})<CR>", channelID)
		},
	)
	return err
}

func (e *FileExplorer) Close() {
	e.active = false
	e.Dirty = false
}

func (e *FileExplorer) mapDirectoryEntries(path string) ([]DirectoryEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirStrings, fileStrings []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirStrings = append(dirStrings, (entry.Name()))
		} else {
			fileStrings = append(fileStrings, (entry.Name()))
		}
	}

	sort.Strings(dirStrings)
	sort.Strings(fileStrings)
	dirs := utils.MapArray(dirStrings, func(dir string) DirectoryEntry {
		return DirectoryEntry{
			ID:        e.nextID(),
			Icon:      "",
			IconClass: "",
			Text:      dir,
			IsDir:     true,
			Path:      path + string(os.PathSeparator) + dir,
		}
	})
	files := utils.MapArray(fileStrings, func(file string) DirectoryEntry {
		return DirectoryEntry{
			ID:        e.nextID(),
			Icon:      "",
			IconClass: "",
			Text:      file,
			Path:      path + string(os.PathSeparator) + file,
			IsDir:     false,
		}
	})
	return append(dirs, files...), nil
}

func (e *FileExplorer) entriesToBufferLines(entries []DirectoryEntry) [][]byte {
	lines := utils.MapArray(entries, func(entry DirectoryEntry) []byte {
		return []byte(strconv.FormatUint(entry.ID, 10))
	})
	if lines == nil {
		return [][]byte{}
	}
	return lines
}

func (e *FileExplorer) setBufferLinesFromEntries(bufNr int, entries []DirectoryEntry) error {
	return e.nvim.SetBufferLines(bufNr, 0, -1, false, e.entriesToBufferLines(entries))
}

func (e *FileExplorer) syncPaneBuffers() error {
	if err := e.setBufferLinesFromEntries(e.parent.BufNr, e.parent.Entries); err != nil {
		return err
	}
	if err := e.setBufferLinesFromEntries(e.current.BufNr, e.current.Entries); err != nil {
		return err
	}
	if err := e.setBufferLinesFromEntries(e.preview.BufNr, e.preview.Entries); err != nil {
		return err
	}
	return nil
}

func (e *FileExplorer) findBufferRowBySelectedEntryID(bufNr int, selectedEntryID uint64) (int, error) {
	lines, err := e.nvim.GetBufferLines(bufNr, 0, -1, false)
	if err != nil {
		return 0, err
	}
	target := strconv.FormatUint(selectedEntryID, 10)
	row, err := utils.FindIndex(lines, func(line []byte) bool {
		return strings.TrimSpace(string(line)) == target
	})
	if err != nil {
		return 0, err
	}
	return row + 1, nil // NOTE: rows are 1 based in neovim
}

func (e *FileExplorer) syncPaneCursorToSelection(winID, bufNr int, selectedEntryID uint64) error {
	if selectedEntryID == 0 {
		return nil
	}
	row, err := e.findBufferRowBySelectedEntryID(bufNr, selectedEntryID)
	if err != nil {
		return err
	}
	return e.nvim.SetWindowCursor(winID, row, 0)
}

func (e *FileExplorer) GoIn() error {

	selectedEntry, err := utils.Find(e.current.Entries, func(entry DirectoryEntry) bool {
		return entry.ID == e.current.SelectedEntryId
	})
	if err != nil {
		return err
	}
	if !selectedEntry.IsDir {
		return fmt.Errorf("cannot go in. selected entry is file: %s", selectedEntry.Path)
	}

	newCurrentPath := selectedEntry.Path

	// Shift panes: current -> parent, selected child directory -> current.
	e.parent.Path = e.current.Path
	e.parent.Entries = e.current.Entries
	e.parent.SelectedEntryId = selectedEntry.ID

	entries, err := e.mapDirectoryEntries(newCurrentPath)
	if err != nil {
		return err
	}
	e.current.Path = newCurrentPath
	e.current.Entries = entries
	if len(entries) > 0 {
		e.current.SelectedEntryId = entries[0].ID
	} else {
		e.current.SelectedEntryId = 0
	}

	e.preview.Entries = []DirectoryEntry{}
	if err := e.syncPaneBuffers(); err != nil {
		return err
	}
	if err := e.syncPaneCursorToSelection(e.parent.WinID, e.parent.BufNr, e.parent.SelectedEntryId); err != nil {
		return err
	}
	if err := e.syncPaneCursorToSelection(e.current.WinID, e.current.BufNr, e.current.SelectedEntryId); err != nil {
		return err
	}
	e.Dirty = true
	return nil
}

func (e *FileExplorer) GoOut() error {
	var err error
	parentPath := e.parent.Path
	newParentDir := path.Dir(parentPath)
	if newParentDir == "." {
		return fmt.Errorf("cannot go out further. already at root dir")
	}
	e.parent.Path = newParentDir
	e.current.Entries = e.parent.Entries
	e.current.SelectedEntryId = e.parent.SelectedEntryId

	entries, err := e.mapDirectoryEntries(e.parent.Path)
	if err != nil {
		return err
	}
	e.parent.Entries = entries
	entry, err := utils.Find(e.parent.Entries, func(entry DirectoryEntry) bool {
		return strings.Contains(e.current.Path, entry.Path)
	})
	if err != nil {
		return err
	}
	e.parent.SelectedEntryId = entry.ID
	e.current.Path = parentPath
	e.preview.Entries = []DirectoryEntry{}
	if err := e.syncPaneBuffers(); err != nil {
		return err
	}
	if err := e.syncPaneCursorToSelection(e.parent.WinID, e.parent.BufNr, e.parent.SelectedEntryId); err != nil {
		return err
	}
	if err := e.syncPaneCursorToSelection(e.current.WinID, e.current.BufNr, e.current.SelectedEntryId); err != nil {
		return err
	}
	e.Dirty = true
	return nil
}
