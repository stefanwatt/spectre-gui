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

//TODO: filesystem operations seem to work. dirty state tracking is broken
// syncing does not set dirty to false for current pane. instead it sets dirty = true for parent and changes all dirs to type file

//TODO: after sync the cursor should be on the same entry as before, even if order has changed through sorting

type FileExplorer struct {
	active             bool
	Dirty              bool
	idCounter          uint64
	nvim               ports.NvimClient
	emitter            ports.UIEmitter
	parent             *Directory
	current            *Directory
	preview            Directory
	openedFromWindow   int
	parentWinID        int
	currentWinID       int
	previewWinID       int
	directoriesByPath  map[string]*Directory
	directoriesByBuf   map[int]*Directory
	dirtyByBuf         map[int]*DirDraft
	sourceByID         map[uint64]DirectoryEntry
	pendingClosePrompt bool
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
		nvim:              nvim,
		idCounter:         1,
		parentWinID:       -1,
		currentWinID:      -1,
		previewWinID:      -1,
		directoriesByPath: map[string]*Directory{},
		directoriesByBuf:  map[int]*Directory{},
		dirtyByBuf:        map[int]*DirDraft{},
		sourceByID:        map[uint64]DirectoryEntry{},
	}
}

func (e *FileExplorer) RegisterHandlers() {
	e.nvim.RegisterHandler("FileExplorerClose", e.onClose)
	e.nvim.RegisterHandler("FileExplorerCursorMoved", e.onCursorMoved)
	e.nvim.RegisterHandler("nvim_buf_lines_event", e.onBufferLines)
	e.nvim.RegisterHandler("nvim_buf_detach_event", e.onBufferDetach)
	e.nvim.RegisterHandler("FileExplorerGoIn", e.onGoIn)
	e.nvim.RegisterHandler("FileExplorerGoOut", e.onGoOut)
	e.nvim.RegisterHandler("FileExplorerSync", e.onSync)
}

func (e *FileExplorer) SetEmitter(emitter ports.UIEmitter) {
	e.emitter = emitter
}

func (e *FileExplorer) Open(filepath *string) error {
	resolvedPath, err := e.resolveOpenFilepath(filepath)
	if err != nil {
		return err
	}
	e.resetSessionState()
	if err := e.createPreviewBuffer(); err != nil {
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
	if err := e.setupCurrentBufferAutocmd(); err != nil {
		return err
	}

	return e.nvim.SetCurrentWindow(e.currentWinID)
}

func (e *FileExplorer) GetParent() Directory {
	if e.parent == nil {
		return Directory{}
	}
	parent := *e.parent
	parent.WinID = e.parentWinID
	return parent
}

func (e *FileExplorer) GetCurrent() Directory {
	if e.current == nil {
		return Directory{}
	}
	current := *e.current
	current.WinID = e.currentWinID
	return current
}

func (e *FileExplorer) GetPreview() Directory {
	preview := e.preview
	preview.WinID = e.previewWinID
	return preview
}

func (e *FileExplorer) GetActive() bool {
	return e.active
}

func (e *FileExplorer) GetCurrentBuf() int {
	if e.current == nil {
		return 0
	}
	return e.current.BufNr
}

func (e *FileExplorer) IsBufDirty(bufNr int) bool {
	if bufNr <= 0 {
		return false
	}
	_, ok := e.dirtyByBuf[bufNr]
	return ok
}

func (e *FileExplorer) UpdateSelectedyEntry(row, col int) error {
	if e.current == nil {
		return fmt.Errorf("[UpdateSelectedyEntry] no current directory")
	}
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

func (e *FileExplorer) UpdateEntries(bufNr, firstline, lastline int, lines []string) error {
	directory, err := e.directoryByBuf(bufNr)
	if err != nil {
		return err
	}
	if !isValidSpliceRange(len(directory.Entries), firstline, lastline) {
		log.Warnf("[UpdateEntries] ignoring out-of-range event buf=%d first=%d last=%d len=%d", bufNr, firstline, lastline, len(directory.Entries))
		return nil
	}
	e.captureDirtyBaseline(bufNr, directory)

	row := firstline
	newEntries := []DirectoryEntry{}
	if len(lines) == 0 {
		// entry deleted
		directory.Entries = append(directory.Entries[:firstline], directory.Entries[lastline:]...)
		e.Dirty = true
		return nil
	}
	for _, line := range lines {
		id, text := parseBufferLine(strings.TrimSpace(string(line)))
		e.Dirty = true
		if id == 0 {
			entry, err := e.findExistingDraftEntry(directory, line, row)
			if err != nil {
				id = e.nextID()
			} else {
				id = entry.ID
			}
			// entirely new entry
			newEntries = append(newEntries, e.createDraftEntry(directory, id, text))
		} else {
			handled := false
			for i, entry := range directory.Entries {
				if entry.ID == id {
					if row+1 == i {
						// renamed / synced replacement at same row
						directory.Entries[i].Text = text
					} else {
						// e.g. copy pasted in same dir
						newEntries = append(newEntries, e.createDraftEntry(directory, e.nextID(), text))
					}
					handled = true
				}
			}
			if !handled {
				// e.g. brought back via undo
				newEntries = append(newEntries, e.createDraftEntry(directory, id, text))
			}
		}
		row += 1
	}
	orig := directory.Entries
	head := append(orig[:firstline:firstline], newEntries...)
	directory.Entries = append(head, orig[lastline:]...)
	return nil
}

func (e *FileExplorer) createDraftEntry(directory *Directory, id uint64, text string) DirectoryEntry {
	return DirectoryEntry{
		ID:        id,
		Text:      text,
		Icon:      FILE_ICON,
		IconClass: "",
		Path:      directory.Path + "/" + text,
		IsDir:     false,
		IsDraft:   true,
	}
}

func (e *FileExplorer) findExistingDraftEntry(directory *Directory, line string, row int) (DirectoryEntry, error) {
	if row > len(directory.Entries)-1 {
		return DirectoryEntry{}, fmt.Errorf("[findExistingDraftEntry] couldnt find candidate: out of bounds")
	}
	candidate := directory.Entries[row]

	if !candidate.IsDraft {
		return DirectoryEntry{}, fmt.Errorf("[findExistingDraftEntry] no draft entry at row=%d", row)
	}
	//NOTE: thinking about soft matching line against candidate.Text, but i cant think of a hard rule
	return candidate, nil
}

func (e *FileExplorer) Close() {
	e.hideClosePrompt()
	e.nvim.Command("tabc")
	e.idCounter = 1
	e.active = false
	e.Dirty = false
	e.resetSessionState()
}

func (e *FileExplorer) GoIn() error {
	if e.current == nil {
		return fmt.Errorf("[GoIn] no current directory")
	}
	selectedEntry, err := utils.Find(e.current.Entries, func(entry DirectoryEntry) bool {
		return entry.ID == e.current.SelectedEntryId
	})
	if err != nil {
		return err
	}
	if !selectedEntry.IsDir {
		return e.OpenFile(selectedEntry.Path)
	}

	newParent := e.current
	newParent.SelectedEntryId = selectedEntry.ID

	newCurrent, err := e.getOrCreateDirectory(selectedEntry.Path)
	if err != nil {
		return err
	}
	if newCurrent.SelectedEntryId == 0 {
		e.selectFirstEntry(newCurrent)
	}

	e.parent = newParent
	e.current = newCurrent

	e.clearPreview()
	if err := e.refreshVisiblePanes(); err != nil {
		return err
	}
	if err := e.setupCurrentBufferAutocmd(); err != nil {
		return err
	}
	e.Dirty = true
	return nil
}

func (e *FileExplorer) GoOut() error {
	if e.parent == nil {
		return fmt.Errorf("[GoOut] no parent directory")
	}
	parentPath := e.parent.Path
	newParentDir := path.Dir(parentPath)
	if newParentDir == "." {
		return fmt.Errorf("cannot go out further. already at root dir")
	}

	newCurrent := e.parent
	newParent, err := e.getOrCreateDirectory(newParentDir)
	if err != nil {
		return err
	}

	entry, err := utils.Find(newParent.Entries, func(entry DirectoryEntry) bool {
		return parentPath == entry.Path
	})
	if err == nil {
		newParent.SelectedEntryId = entry.ID
	}

	e.current = newCurrent
	e.parent = newParent

	e.clearPreview()
	if err := e.refreshVisiblePanes(); err != nil {
		return err
	}
	if err := e.setupCurrentBufferAutocmd(); err != nil {
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
	return nil
}

func (e *FileExplorer) initializePaneDirectories(filepath string) error {
	currentPath := path.Dir(filepath)
	parentPath := path.Dir(currentPath)
	if parentPath == "." {
		return fmt.Errorf("[FileExplorer] parent dir doesnt exist")
	}

	parentDir, err := e.getOrCreateDirectory(parentPath)
	if err != nil {
		return err
	}
	currentDir, err := e.getOrCreateDirectory(currentPath)
	if err != nil {
		return err
	}
	e.parent = parentDir
	e.current = currentDir
	return nil
}

func (e *FileExplorer) loadDirectory(directory *Directory, directoryPath string) error {
	entries, err := e.mapDirectoryEntries(directoryPath)
	if err != nil {
		return err
	}
	directory.Path = directoryPath
	directory.Entries = entries
	e.indexSourceEntries(entries)
	return nil
}

func (e *FileExplorer) openPaneWindows() error {
	if e.parent == nil || e.current == nil {
		return fmt.Errorf("[openPaneWindows] parent/current not initialized")
	}
	if err := e.nvim.Command("tab new"); err != nil {
		return err
	}

	winID, err := e.nvim.CurrentWindow()
	if err != nil {
		return err
	}
	e.parentWinID = winID

	if err := e.nvim.SetBufferToWindow(e.parentWinID, e.parent.BufNr); err != nil {
		return err
	}
	if err := e.nvim.OpenSplitRight(&e.currentWinID, e.current.BufNr); err != nil {
		return err
	}
	if err := e.nvim.OpenSplitRight(&e.previewWinID, e.preview.BufNr); err != nil {
		return err
	}
	return nil
}

func (e *FileExplorer) refreshVisiblePanes() error {
	if e.parent == nil || e.current == nil {
		return fmt.Errorf("[refreshVisiblePanes] parent/current not initialized")
	}
	if err := e.nvim.SetBufferToWindow(e.parentWinID, e.parent.BufNr); err != nil {
		return err
	}
	if err := e.nvim.SetBufferToWindow(e.currentWinID, e.current.BufNr); err != nil {
		return err
	}
	if err := e.nvim.SetBufferToWindow(e.previewWinID, e.preview.BufNr); err != nil {
		return err
	}
	if err := e.syncPaneBuffers(); err != nil {
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
	if e.parent == nil || e.current == nil {
		return fmt.Errorf("[updateSelectedEntries] parent/current not initialized")
	}
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

func (e *FileExplorer) setupKeymaps(bufNr int) error {
	err := e.nvim.CreateBufferKeymap(
		bufNr,
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
		bufNr,
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
		bufNr,
		"n",
		"<Left>",
		func(channelID int) string {
			return fmt.Sprintf(":lua vim.rpcnotify(%d, 'FileExplorerGoOut', {})<CR>", channelID)
		},
	)
	if err != nil {
		return err
	}

	err = e.nvim.CreateBufferKeymap(
		bufNr,
		"n",
		"=",
		func(channelID int) string {
			return fmt.Sprintf(":lua vim.rpcnotify(%d, 'FileExplorerSync', {})<CR>", channelID)
		},
	)
	return err
}

func (e *FileExplorer) setupCurrentBufferAutocmd() error {
	if e.current == nil {
		return fmt.Errorf("[setupCurrentBufferAutocmd] no current directory")
	}
	return e.nvim.CreateBufferAutocmd(e.currentWinID, e.current.BufNr, `
		print("chanid=" .. tostring(chan_id) .. " winId=" .. tostring(winId) .. " bufNr=" .. tostring(bufNr))
		if (args.buf ~= bufNr)then
			return 
		end
		local cursor = vim.api.nvim_win_get_cursor(0)
		vim.rpcnotify(chan_id, "FileExplorerCursorMoved", cursor)
	`)
}

func (e *FileExplorer) resetSessionState() {
	e.parent = nil
	e.current = nil
	e.preview = Directory{}
	e.parentWinID = -1
	e.currentWinID = -1
	e.previewWinID = -1
	e.directoriesByPath = map[string]*Directory{}
	e.directoriesByBuf = map[int]*Directory{}
	e.dirtyByBuf = map[int]*DirDraft{}
	e.sourceByID = map[uint64]DirectoryEntry{}
	e.pendingClosePrompt = false
}

func (e *FileExplorer) createPreviewBuffer() error {
	bufNr, err := e.nvim.CreateBuffer(true, false)
	if err != nil {
		return err
	}
	e.preview = Directory{BufNr: bufNr, Entries: []DirectoryEntry{}}
	return nil
}

func (e *FileExplorer) getOrCreateDirectory(directoryPath string) (*Directory, error) {
	if directory, ok := e.directoriesByPath[directoryPath]; ok {
		return directory, nil
	}

	bufNr, err := e.nvim.CreateBuffer(true, false)
	if err != nil {
		return nil, err
	}
	directory := &Directory{BufNr: bufNr, Path: directoryPath}
	if err := e.loadDirectory(directory, directoryPath); err != nil {
		return nil, err
	}
	e.selectFirstEntry(directory)
	if err := e.setBufferLinesFromEntries(directory.BufNr, directory.Entries); err != nil {
		return nil, err
	}
	if err := e.setupKeymaps(directory.BufNr); err != nil {
		return nil, err
	}

	var ok bool
	if ok, err = e.nvim.AttachBuffer(directory.BufNr, false, map[string]any{}); err != nil {
		return nil, err
	}
	if !ok {
		log.Errorf("could not attach to buffer %d", directory.BufNr)
	}

	e.directoriesByPath[directoryPath] = directory
	e.directoriesByBuf[directory.BufNr] = directory
	return directory, nil
}

func (e *FileExplorer) directoryByBuf(bufNr int) (*Directory, error) {
	directory, ok := e.directoriesByBuf[bufNr]
	if !ok {
		return nil, fmt.Errorf("unknown directory buffer %d", bufNr)
	}
	return directory, nil
}

func isValidSpliceRange(entriesLen, firstline, lastline int) bool {
	if firstline < 0 || lastline < 0 || firstline > lastline {
		return false
	}
	if firstline > entriesLen || lastline > entriesLen {
		return false
	}
	return true
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
	if e.preview.BufNr == 0 {
		return nil
	}
	if err := e.setBufferLinesFromEntries(e.preview.BufNr, e.preview.Entries); err != nil {
		return err
	}
	return nil
}
