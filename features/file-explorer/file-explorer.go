package fileexplorer

import (
	"bytes"
	"fmt"
	"os"
	path "path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"nvim-gui/core/ports"
	"nvim-gui/utils"

	"github.com/charmbracelet/log"
)

var (
	FILE_ICON = ""
	DIR_ICON  = ""
)

//TODO: bug: switching mode does not send fileexplorer update apparently
// cause cursor shape doesnt update until you move

//TODO: use if init syntax wherever possible & reasonable

//TODO: buf attach mechanism doesnt create new row and doesnt work when id breaks

//TODO: this file is getting a bit long. find opportunities to modularize

//TODO: check if we should use more pointers

//TODO: use uint64 where possible & reasonable

type FileExplorer struct {
	active           bool
	Dirty            bool
	idCounter        uint64
	nvim             ports.NvimClient
	parent           Directory
	current          Directory
	preview          Directory
	openedFromWindow int
}

type DirectoryEntry struct {
	ID        uint64 `json:"id"`
	Icon      string `json:"icon"`
	IconClass string `json:"iconClass"`
	Text      string `json:"text"`
	Path      string `json:"path"`
	IsDir     bool   `json:"isDir"`
	IsDraft   bool
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
		nvim:      nvim,
		idCounter: 1,
	}
}

func (e *FileExplorer) RegisterHandlers() {
	e.nvim.RegisterHandler("FileExplorerClose", e.onClose)
	e.nvim.RegisterHandler("FileExplorerCursorMoved", e.onCursorMoved)
	e.nvim.RegisterHandler("nvim_buf_lines_event", e.onBufferLines)
	e.nvim.RegisterHandler("nvim_buf_detach_event", e.onBufferDetach)
	e.nvim.RegisterHandler("FileExplorerGoIn", e.onGoIn)
	e.nvim.RegisterHandler("FileExplorerGoOut", e.onGoOut)
}

func (e *FileExplorer) Open(filepath *string) error {
	resolvedPath, err := e.resolveOpenFilepath(filepath)
	if err != nil {
		return err
	}
	if err := e.createPaneBuffers(); err != nil {
		return err
	}
	if err := e.initializePaneDirectories(resolvedPath); err != nil {
		return err
	}
	if err := e.openPaneWindows(); err != nil {
		return err
	}
	if err := e.updateSelectedEntries(resolvedPath); err != nil {
		return err
	}

	e.clearPreview()
	if err := e.refreshVisiblePanes(); err != nil {
		return err
	}

	e.active = true
	e.Dirty = true
	if err := e.setupKeymaps(); err != nil {
		return err
	}
	if err := e.setupCurrentBufferAutocmd(); err != nil {
		return err
	}

	var ok bool
	if ok, err = e.nvim.AttachBuffer(e.current.BufNr, false, map[string]any{}); err != nil {
		return err
	}
	if !ok {
		log.Error("could not attach to buffer")
	}

	return e.nvim.SetCurrentWindow(e.current.WinID)
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

func (e *FileExplorer) UpdateSelectedyEntry(row, col int) error {
	if row < 0 || col < 0 || row >= len(e.current.Entries) {
		return fmt.Errorf("[UpdateSelectedyEntry] row/col out of bounds")
	}

	lines, err := e.nvim.GetBufferLines(e.current.BufNr, row, row+1, false)
	if err != nil {
		return fmt.Errorf("[UpdateSelectedyEntry] get buffer line failed: %w", err)
	}
	if len(lines) == 0 {
		return fmt.Errorf("[UpdateSelectedyEntry] no line at row=%d", row)
	}

	id := e.current.Entries[row].ID
	cursorChanged := e.current.CursorCol != col
	selectionChanged := e.current.SelectedEntryId != id
	//TODO: cursor position should be entirely handled in Go not in svelte
	// its currently broken for draft entries
	e.current.CursorCol = col
	e.current.SelectedEntryId = id
	if cursorChanged || selectionChanged {
		e.Dirty = true
	}
	return nil
}

func (e *FileExplorer) UpdateEntries(firstline, lastline int, lines []string) error {
	row := firstline 
	newEntries := []DirectoryEntry{}
	if len(lines) == 0 {
		// entry deleted
		e.current.Entries = append(e.current.Entries[:firstline], e.current.Entries[lastline:]...)
		return nil
	}
	for _, line := range lines {
		id, text := parseBufferLine(strings.TrimSpace(string(line)))
		e.Dirty = true
		if id == 0 {
			entry, err := e.findExistingDraftEntry(line, row)
			if err != nil {
				id = e.nextID()
			} else {
				id = entry.ID
			}
			//entirely new entry
			newEntries = append(newEntries, e.createDraftEntry(id, text))
		} else {
			handled := false
			for i, entry := range e.current.Entries {
				if entry.ID == id {
					if row+1 == i {
						// renamed
						e.current.Entries[i].Text = text
					} else {
						// e.g. copy pasted in same dir
						newEntries = append(newEntries, e.createDraftEntry(e.nextID(), text))
					}
					handled = true
				}
			}
			if !handled {
				// e.g. brought back via undo
				newEntries = append(newEntries, e.createDraftEntry(id, text))
			}
		}
		row += 1
	}
	orig := e.current.Entries
	head := append(orig[:firstline:firstline], newEntries...)
	e.current.Entries = append(head, orig[lastline:]...)
	return nil
}

func (e *FileExplorer) createDraftEntry(id uint64, text string) DirectoryEntry {
	return DirectoryEntry{
		ID:        id,
		Text:      text,
		Icon:      FILE_ICON,
		IconClass: "",
		Path:      e.current.Path + "/" + text,
		IsDir:     false,
		IsDraft:   true,
	}
}

func (e *FileExplorer) findExistingDraftEntry(line string, row int) (DirectoryEntry, error) {
	if row > len(e.current.Entries)-1 {
		return DirectoryEntry{}, fmt.Errorf("[findExistingDraftEntry] couldnt find candidate: out of bounds")
	}
	candidate := e.current.Entries[row]

	if !candidate.IsDraft {
		return DirectoryEntry{}, fmt.Errorf("[findExistingDraftEntry] no draft entry at row=%d", row)
	}
	//NOTE: thinking about soft matching line against candidate.Text, but i cant think of a hard rule
	return candidate, nil
}

func (e *FileExplorer) Close() {
	e.nvim.Command("tabc")
	e.idCounter = 1
	e.active = false
	e.Dirty = false
}

func (e *FileExplorer) GoIn() error {
	selectedEntry, err := utils.Find(e.current.Entries, func(entry DirectoryEntry) bool {
		return entry.ID == e.current.SelectedEntryId
	})
	if err != nil {
		return err
	}
	if !selectedEntry.IsDir {
		e.OpenFile(selectedEntry.Path)
	}

	newCurrentPath := selectedEntry.Path

	// Shift panes: current -> parent, selected child directory -> current.
	e.parent.Path = e.current.Path
	e.parent.Entries = e.current.Entries
	e.parent.SelectedEntryId = selectedEntry.ID

	if err := e.loadDirectory(&e.current, newCurrentPath); err != nil {
		return err
	}
	e.selectFirstEntry(&e.current)

	e.clearPreview()
	if err := e.refreshVisiblePanes(); err != nil {
		return err
	}
	e.Dirty = true
	return nil
}

func (e *FileExplorer) GoOut() error {
	parentPath := e.parent.Path
	newParentDir := path.Dir(parentPath)
	if newParentDir == "." {
		return fmt.Errorf("cannot go out further. already at root dir")
	}

	e.parent.Path = newParentDir
	e.current.Entries = e.parent.Entries
	e.current.SelectedEntryId = e.parent.SelectedEntryId

	if err := e.loadDirectory(&e.parent, e.parent.Path); err != nil {
		return err
	}
	entry, err := utils.Find(e.parent.Entries, func(entry DirectoryEntry) bool {
		return strings.Contains(e.current.Path, entry.Path)
	})
	if err != nil {
		return err
	}

	e.parent.SelectedEntryId = entry.ID
	e.current.Path = parentPath

	e.clearPreview()
	if err := e.refreshVisiblePanes(); err != nil {
		return err
	}
	e.Dirty = true
	return nil
}

func (e *FileExplorer) OpenFile(filepath string) error {
	e.Close()
	err := e.nvim.SetCurrentWindow(e.openedFromWindow)
	if err != nil {
		return err
	}
	err = e.nvim.Command("e " + filepath)
	return err
}

func (e *FileExplorer) nextID() uint64 {
	return atomic.AddUint64(&e.idCounter, 1)
}

func (e *FileExplorer) resolveOpenFilepath(filepath *string) (string, error) {
	var err error
	var resolvedPath string
	if filepath == nil {
		resolvedPath, err = e.nvim.GetCurrentFilepath()
	}
	if err != nil {
		return "", err
	}
	return resolvedPath, nil
}

func (e *FileExplorer) createPaneBuffers() error {
	for _, directory := range []*Directory{&e.parent, &e.current, &e.preview} {
		bufNr, err := e.nvim.CreateBuffer(true, false)
		if err != nil {
			return err
		}
		directory.BufNr = bufNr
	}
	return nil
}

func (e *FileExplorer) initializePaneDirectories(filepath string) error {
	currentPath := path.Dir(filepath)
	parentPath := path.Dir(currentPath)
	if parentPath == "." {
		return fmt.Errorf("[FileExplorer] parent dir doesnt exist")
	}

	if err := e.loadDirectory(&e.parent, parentPath); err != nil {
		return err
	}
	if err := e.loadDirectory(&e.current, currentPath); err != nil {
		return err
	}
	return nil
}

func (e *FileExplorer) loadDirectory(directory *Directory, directoryPath string) error {
	entries, err := e.mapDirectoryEntries(directoryPath)
	if err != nil {
		return err
	}
	directory.Path = directoryPath
	directory.Entries = entries
	return nil
}

func (e *FileExplorer) openPaneWindows() error {
	if err := e.nvim.Command("tab new"); err != nil {
		return err
	}

	winID, err := e.nvim.CurrentWindow()
	if err != nil {
		return err
	}
	e.parent.WinID = winID

	if err := e.nvim.SetBufferToWindow(e.parent.WinID, e.parent.BufNr); err != nil {
		return err
	}
	if err := e.nvim.OpenSplitRight(&e.current.WinID, e.current.BufNr); err != nil {
		return err
	}
	if err := e.nvim.OpenSplitRight(&e.preview.WinID, e.preview.BufNr); err != nil {
		return err
	}
	return nil
}

func (e *FileExplorer) refreshVisiblePanes() error {
	if err := e.syncPaneBuffers(); err != nil {
		return err
	}
	if err := e.syncPaneCursorToSelection(e.parent.WinID, e.parent.BufNr, e.parent.SelectedEntryId); err != nil {
		return err
	}
	if err := e.syncPaneCursorToSelection(e.current.WinID, e.current.BufNr, e.current.SelectedEntryId); err != nil {
		return err
	}
	return nil
}

func (e *FileExplorer) clearPreview() {
	e.preview.Entries = []DirectoryEntry{}
}

func (e *FileExplorer) selectFirstEntry(directory *Directory) {
	if len(directory.Entries) > 0 {
		directory.SelectedEntryId = directory.Entries[0].ID
		return
	}
	directory.SelectedEntryId = 0
}

func (e *FileExplorer) updateSelectedEntries(filepath string) error {
	entry, err := utils.Find(e.parent.Entries, func(entry DirectoryEntry) bool {
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

func (e *FileExplorer) setupCurrentBufferAutocmd() error {
	return e.nvim.CreateBufferAutocmd(e.current.WinID, e.current.BufNr, `
		print("chanid=" .. tostring(chan_id) .. " winId=" .. tostring(winId) .. " bufNr=" .. tostring(bufNr))
		if (args.buf ~= bufNr)then
			return 
		end
		local cursor = vim.api.nvim_win_get_cursor(0)
		vim.rpcnotify(chan_id, "FileExplorerCursorMoved", cursor)
	`)
}

func parseBufferLine(line string) (uint64, string) {
	if line == "" {
		return 0, ""
	}
	parts := strings.SplitN(line, "/", 2)
	idPart := strings.TrimSpace(parts[0])
	if idPart == "" {
		return 0, line
	}
	id, err := strconv.ParseUint(idPart, 10, 64)
	if err != nil {
		return 0, line
	}

	text := ""
	if len(parts) == 2 {
		text = parts[1]
	}

	return id, text
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
			dirStrings = append(dirStrings, entry.Name())
		} else {
			fileStrings = append(fileStrings, entry.Name())
		}
	}

	sort.Slice(dirStrings, func(i, j int) bool {
		return strings.ToLower(dirStrings[i]) < strings.ToLower(dirStrings[j])
	})
	sort.Slice(fileStrings, func(i, j int) bool {
		return strings.ToLower(fileStrings[i]) < strings.ToLower(fileStrings[j])
	})

	dirs := utils.MapArray(dirStrings, func(dir string) DirectoryEntry {
		return DirectoryEntry{
			ID:        e.nextID(),
			Icon:      DIR_ICON,
			IconClass: "",
			Text:      dir,
			IsDir:     true,
			Path:      path + string(os.PathSeparator) + dir,
		}
	})
	files := utils.MapArray(fileStrings, func(file string) DirectoryEntry {
		return DirectoryEntry{
			ID:        e.nextID(),
			Icon:      FILE_ICON,
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
		return fmt.Appendf(nil, "%d/%s", entry.ID, entry.Text)
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
	if _, err := e.entriesForBuffer(bufNr); err != nil {
		return 0, err
	}
	lines, err := e.nvim.GetBufferLines(bufNr, 0, -1, false)
	if err != nil {
		return 0, err
	}
	row, err := utils.FindIndex(lines, func(line []byte) bool {
		id, _ := parseBufferLine(strings.TrimSpace(string(line)))
		return id == selectedEntryID
	})
	if err != nil {
		return 0, err
	}
	return row + 1, nil // NOTE: rows are 1 based in neovim
}

func (e *FileExplorer) entriesForBuffer(bufNr int) ([]DirectoryEntry, error) {
	switch bufNr {
	case e.parent.BufNr:
		return e.parent.Entries, nil
	case e.current.BufNr:
		return e.current.Entries, nil
	case e.preview.BufNr:
		return e.preview.Entries, nil
	default:
		return nil, fmt.Errorf("unknown buffer %d", bufNr)
	}
}

func (e *FileExplorer) syncPaneCursorToSelection(winID, bufNr int, selectedEntryID uint64) error {
	if selectedEntryID == 0 {
		return nil
	}

	var err error
	var lines [][]byte
	if lines, err = e.nvim.GetBufferLines(bufNr, 0, -1, true); err != nil {
		return err
	}
	s := string(bytes.Join(lines, []byte("\n")))
	log.Debug(s)
	var row int
	if row, err = e.findBufferRowBySelectedEntryID(bufNr, selectedEntryID); err != nil {
		return err
	}
	return e.nvim.SetWindowCursor(winID, row, 0)
}
