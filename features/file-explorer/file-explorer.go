package fileexplorer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	path "path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"nvim-gui/core/ports"
	"nvim-gui/utils"

	"github.com/charmbracelet/log"
)

var (
	FILE_ICON = ""
	DIR_ICON  = ""
)

type PreviewKind string

const (
	PreviewKindNone       PreviewKind = ""
	PreviewKindDirectory  PreviewKind = "directory"
	PreviewKindTextFile   PreviewKind = "textFile"
	PreviewKindLocalImage PreviewKind = "localImage"
)

const (
	previewDebounce      = 75 * time.Millisecond
	previewBinaryProbe   = 1024
	previewHugeFileBytes = 1024 * 1024
	defaultPreviewRows   = 32
	previewCellWidthPx   = 12
	previewCellHeightPx  = 28
)

var previewImageExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".bmp":  true,
	".ico":  true,
}

//TODO: bug: switching mode does not send fileexplorer update apparently
// cause cursor shape doesnt update until you move

//TODO: use if init syntax wherever possible & reasonable

//TODO: buf attach mechanism doesnt create new row and doesnt work when id breaks

//TODO: this file is getting a bit long. find opportunities to modularize

//TODO: check if we should use more pointers

//TODO: use uint64 where possible & reasonable

type FileExplorer struct {
	active             bool
	Dirty              bool
	idCounter          uint64
	nvim               ports.NvimClient
	emitter            ports.UIEmitter
	parent             *Directory
	current            *Directory
	preview            Directory
	previewKind        PreviewKind
	previewPath        string
	previewTitle       string
	previewFiletype    string
	previewTextLines   []string
	previewTextBufNr   int
	previewTextBufPath string
	previewGeneration  uint64
	previewTimer       *time.Timer
	previewRows        int
	openedFromWindow   int
	parentWinID        int
	currentWinID       int
	previewWinID       int
	directoriesByPath  map[string]*Directory
	directoriesByBuf   map[int]*Directory
	dirtyByBuf         map[int]*DirDraft
	sourceByID         map[uint64]DirectoryEntry
	idByPath           map[string]uint64
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
		previewRows:       defaultPreviewRows,
		directoriesByPath: map[string]*Directory{},
		directoriesByBuf:  map[int]*Directory{},
		dirtyByBuf:        map[int]*DirDraft{},
		sourceByID:        map[uint64]DirectoryEntry{},
		idByPath:          map[string]uint64{},
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
	e.stopPreviewTimer()
	e.deletePreviewTextBuffer()
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
	if err := e.refreshPreviewForSelectedEntry(); err != nil {
		log.Warnf("[FileExplorer] initial preview failed: %v", err)
	}
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

func (e *FileExplorer) GetPreviewKind() PreviewKind {
	return e.previewKind
}

func (e *FileExplorer) GetPreviewPath() string {
	return e.previewPath
}

func (e *FileExplorer) GetPreviewTitle() string {
	return e.previewTitle
}

func (e *FileExplorer) GetPreviewFiletype() string {
	return e.previewFiletype
}

func (e *FileExplorer) GetPreviewTextLines() []string {
	return append([]string(nil), e.previewTextLines...)
}

func (e *FileExplorer) GetPreviewWinID() int {
	return e.previewWinID
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

	line, err := e.getBufferLine(e.current.BufNr, row)
	if err != nil {
		return fmt.Errorf("[UpdateSelectedyEntry] get buffer line failed: %w", err)
	}

	id := e.current.Entries[row].ID
	selectionChanged := e.current.SelectedEntryId != id

	visibleCol := bufferColToVisibleCol(line, col)
	correctedRawCol := col
	if selectionChanged && e.current.SelectedEntryId != 0 {
		visibleCol = clampVisibleCol(line, e.current.CursorCol)
		correctedRawCol = visibleColToBufferCol(line, visibleCol)
	} else if col < bufferLineVisibleStartCol(line) {
		visibleCol = 0
		correctedRawCol = visibleColToBufferCol(line, visibleCol)
	}

	cursorChanged := e.current.CursorCol != visibleCol
	e.current.CursorCol = visibleCol
	e.current.SelectedEntryId = id
	if cursorChanged || selectionChanged {
		e.Dirty = true
	}
	if selectionChanged {
		e.schedulePreviewRefresh(e.current.Entries[row])
	}
	if correctedRawCol != col && e.currentWinID >= 0 {
		return e.nvim.SetWindowCursor(e.currentWinID, row+1, correctedRawCol)
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
				if entry.ID != id {
					continue
				}
				if row == i {
					// renamed / synced replacement at same row
					entry.Text = text
					entry.Path = path.Join(directory.Path, text)
					newEntries = append(newEntries, entry)
				} else {
					// e.g. copy pasted in same dir
					newEntries = append(newEntries, e.createDraftEntry(directory, e.nextID(), text))
				}
				handled = true
				break
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
	entryPath := path.Join(directory.Path, text)
	if e.idByPath == nil {
		e.idByPath = map[string]uint64{}
	}
	if _, exists := e.idByPath[entryPath]; !exists {
		e.idByPath[entryPath] = id
	}
	return DirectoryEntry{
		ID:        id,
		Text:      text,
		Icon:      FILE_ICON,
		IconClass: "",
		Path:      entryPath,
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
	e.stopPreviewTimer()
	e.deletePreviewTextBuffer()
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
	if err := e.refreshPreviewForSelectedEntry(); err != nil {
		log.Warnf("[FileExplorer] preview refresh failed: %v", err)
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
	if err := e.refreshPreviewForSelectedEntry(); err != nil {
		log.Warnf("[FileExplorer] preview refresh failed: %v", err)
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
	if err := e.syncVisiblePaneCursorsToSelection(); err != nil {
		return err
	}
	return nil
}

func (e *FileExplorer) syncVisiblePaneCursorsToSelection() error {
	if err := e.syncPaneCursorToSelection(e.parentWinID, e.parent); err != nil {
		return err
	}
	return e.syncPaneCursorToSelection(e.currentWinID, e.current)
}

func (e *FileExplorer) syncPaneCursorToSelection(winID int, directory *Directory) error {
	if directory == nil || winID < 0 || directory.SelectedEntryId == 0 {
		return nil
	}
	row, ok, err := e.findBufferRowBySelectedEntryID(directory.BufNr, directory.SelectedEntryId)
	if err != nil {
		return err
	}
	if !ok {
		row, ok = findDirectoryRowBySelectedEntryID(directory)
	}
	if !ok {
		log.Warnf("[FileExplorer] selected entry %d not found in buffer %d", directory.SelectedEntryId, directory.BufNr)
		return nil
	}
	line, err := e.getBufferLine(directory.BufNr, row-1)
	if err != nil {
		return err
	}
	return e.nvim.SetWindowCursor(winID, row, visibleColToBufferCol(line, directory.CursorCol))
}

func (e *FileExplorer) findBufferRowBySelectedEntryID(bufNr int, selectedEntryID uint64) (int, bool, error) {
	lines, err := e.nvim.GetBufferLines(bufNr, 0, -1, false)
	if err != nil {
		return 0, false, err
	}
	for i, line := range lines {
		id, _ := parseBufferLine(strings.TrimSpace(string(line)))
		if id == selectedEntryID {
			return i + 1, true, nil
		}
	}
	return 0, false, nil
}

func (e *FileExplorer) getBufferLine(bufNr, row int) (string, error) {
	lines, err := e.nvim.GetBufferLines(bufNr, row, row+1, false)
	if err != nil {
		return "", err
	}
	if len(lines) == 0 {
		return "", fmt.Errorf("no line at row=%d", row)
	}
	return string(lines[0]), nil
}

func findDirectoryRowBySelectedEntryID(directory *Directory) (int, bool) {
	if directory == nil || directory.SelectedEntryId == 0 {
		return 0, false
	}
	for i, entry := range directory.Entries {
		if entry.ID == directory.SelectedEntryId {
			return i + 1, true
		}
	}
	return 0, false
}

func bufferLineVisibleStartCol(line string) int {
	if line == "" || line[0] < '0' || line[0] > '9' {
		return 0
	}
	for i := 1; i < len(line); i++ {
		if line[i] >= '0' && line[i] <= '9' {
			continue
		}
		if line[i] == '/' {
			return i + 1
		}
		return 0
	}
	return 0
}

func bufferLineVisibleText(line string) string {
	return line[bufferLineVisibleStartCol(line):]
}

func bufferColToVisibleCol(line string, rawCol int) int {
	start := bufferLineVisibleStartCol(line)
	return clampInt(rawCol-start, 0, len(bufferLineVisibleText(line)))
}

func visibleColToBufferCol(line string, visibleCol int) int {
	start := bufferLineVisibleStartCol(line)
	return start + clampVisibleCol(line, visibleCol)
}

func clampVisibleCol(line string, visibleCol int) int {
	return clampInt(visibleCol, 0, len(bufferLineVisibleText(line)))
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (e *FileExplorer) clearPreview() {
	e.preview.Entries = []DirectoryEntry{}
	e.previewKind = PreviewKindNone
	e.previewPath = ""
	e.previewTitle = ""
	e.previewFiletype = ""
	e.previewTextLines = nil
}

func (e *FileExplorer) stopPreviewTimer() {
	e.previewGeneration++
	if e.previewTimer == nil {
		return
	}
	e.previewTimer.Stop()
	e.previewTimer = nil
}

func (e *FileExplorer) selectedEntry() (DirectoryEntry, bool) {
	if e.current == nil || e.current.SelectedEntryId == 0 {
		return DirectoryEntry{}, false
	}
	for _, entry := range e.current.Entries {
		if entry.ID == e.current.SelectedEntryId {
			return entry, true
		}
	}
	return DirectoryEntry{}, false
}

func (e *FileExplorer) schedulePreviewRefresh(entry DirectoryEntry) {
	e.previewGeneration++
	generation := e.previewGeneration
	if e.previewTimer != nil {
		e.previewTimer.Stop()
	}
	e.previewTimer = time.AfterFunc(previewDebounce, func() {
		if !e.active || generation != e.previewGeneration {
			return
		}
		if err := e.refreshPreviewForEntry(entry, generation); err != nil {
			log.Warnf("[FileExplorer] preview refresh failed path=%s err=%v", entry.Path, err)
		}
	})
}

func (e *FileExplorer) refreshPreviewForSelectedEntry() error {
	entry, ok := e.selectedEntry()
	if !ok {
		e.clearPreview()
		return nil
	}
	e.previewGeneration++
	return e.refreshPreviewForEntry(entry, e.previewGeneration)
}

func (e *FileExplorer) refreshPreviewForEntry(entry DirectoryEntry, generation uint64) error {
	if !e.active || generation != e.previewGeneration {
		return nil
	}
	if entry.Path == "" {
		e.clearPreview()
		e.Dirty = true
		return nil
	}

	kind := PreviewKindTextFile
	if entry.IsDir {
		kind = PreviewKindDirectory
	} else if isPreviewImage(entry.Path) {
		kind = PreviewKindLocalImage
	}
	e.beginPreviewTransition(entry, kind)

	var err error
	switch kind {
	case PreviewKindDirectory:
		err = e.refreshDirectoryPreview(entry)
	case PreviewKindLocalImage:
		err = e.refreshImagePreview(entry)
	case PreviewKindTextFile:
		err = e.refreshTextPreview(entry)
	}
	if err != nil {
		e.restoreCurrentPaneFocus()
		e.Dirty = true
		e.requestRedraw()
	}
	return err
}

func (e *FileExplorer) beginPreviewTransition(entry DirectoryEntry, kind PreviewKind) {
	previewBufNr := e.preview.BufNr
	e.preview = Directory{
		WinID:   e.previewWinID,
		BufNr:   previewBufNr,
		Entries: []DirectoryEntry{},
		Path:    entry.Path,
	}
	e.previewKind = kind
	e.previewPath = entry.Path
	e.previewTitle = previewEntryTitle(entry)
	e.previewFiletype = ""
	e.previewTextLines = nil
	e.Dirty = true
}

func (e *FileExplorer) refreshDirectoryPreview(entry DirectoryEntry) error {
	entries, err := e.mapDirectoryEntries(entry.Path)
	if err != nil {
		return err
	}
	previewBufNr := e.preview.BufNr
	e.preview = Directory{
		WinID:     e.previewWinID,
		BufNr:     previewBufNr,
		Entries:   entries,
		Path:      entry.Path,
		CursorCol: 0,
	}
	e.previewKind = PreviewKindDirectory
	e.previewPath = entry.Path
	e.previewTitle = previewEntryTitle(entry)
	e.previewFiletype = ""
	e.previewTextLines = nil
	e.restoreCurrentPaneFocus()
	e.Dirty = true
	e.requestRedraw()
	return nil
}

func (e *FileExplorer) refreshImagePreview(entry DirectoryEntry) error {
	e.previewKind = PreviewKindLocalImage
	e.previewPath = entry.Path
	e.previewTitle = previewEntryTitle(entry)
	e.previewFiletype = ""
	e.previewTextLines = nil
	e.preview.Entries = []DirectoryEntry{}
	e.restoreCurrentPaneFocus()
	e.Dirty = true
	e.requestRedraw()
	return nil
}

func (e *FileExplorer) refreshTextPreview(entry DirectoryEntry) error {
	rows := e.previewRows
	if rows <= 0 {
		rows = defaultPreviewRows
	}
	lines, binary, huge, err := readPreviewLines(entry.Path, rows)
	if err != nil {
		lines = []string{fmt.Sprintf("[preview unavailable: %v]", err)}
	}
	if binary {
		lines = []string{"[binary file preview unavailable]"}
	}
	e.previewTextLines = append([]string(nil), lines...)
	bufNr, err := e.ensurePreviewTextBuffer(entry.Path)
	if err != nil {
		return err
	}
	if err := e.nvim.SetBufferOption(bufNr, "modifiable", true); err != nil {
		return err
	}
	if err := e.nvim.SetBufferLines(bufNr, 0, -1, false, stringsToBytes(lines)); err != nil {
		return err
	}
	filetype, err := e.configurePreviewTextBuffer(bufNr, entry.Path, huge || binary)
	if err != nil {
		return err
	}
	if e.previewWinID >= 0 {
		if err := e.nvim.SetBufferToWindow(e.previewWinID, bufNr); err != nil {
			return err
		}
		e.positionPreviewWindowAtTop()
	}
	e.previewKind = PreviewKindTextFile
	e.previewPath = entry.Path
	e.previewTitle = previewEntryTitle(entry)
	e.previewFiletype = filetype
	e.previewTextLines = append([]string(nil), lines...)
	e.preview.Entries = []DirectoryEntry{}
	e.restoreCurrentPaneFocus()
	e.Dirty = true
	e.requestRedraw()
	return nil
}

func (e *FileExplorer) ensurePreviewTextBuffer(previewPath string) (int, error) {
	if e.previewTextBufNr > 0 {
		e.previewTextBufPath = previewPath
		return e.previewTextBufNr, nil
	}
	bufNr, err := e.nvim.CreateBuffer(false, true)
	if err != nil {
		return 0, err
	}
	e.previewTextBufNr = bufNr
	e.previewTextBufPath = previewPath
	return bufNr, nil
}

func (e *FileExplorer) deletePreviewTextBuffer() {
	if e.previewTextBufNr <= 0 {
		e.previewTextBufPath = ""
		return
	}
	if err := e.nvim.DeleteBuffer(e.previewTextBufNr, true); err != nil {
		log.Warnf("[FileExplorer] delete preview buffer failed buf=%d err=%v", e.previewTextBufNr, err)
	}
	e.previewTextBufNr = 0
	e.previewTextBufPath = ""
}

func (e *FileExplorer) configurePreviewTextBuffer(bufNr int, filename string, skipHighlight bool) (string, error) {
	var filetype string
	err := e.nvim.ExecLua(`
		local bufnr, filename, skip_highlight = ...
		if not vim.api.nvim_buf_is_valid(bufnr) then
			return ""
		end
		pcall(vim.api.nvim_buf_set_name, bufnr, "nvim-gui-preview://" .. filename)
		vim.bo[bufnr].buftype = "nofile"
		vim.bo[bufnr].bufhidden = "wipe"
		vim.bo[bufnr].swapfile = false
		vim.bo[bufnr].readonly = true
		local ft = ""
		if not skip_highlight then
			local ok, detected = pcall(vim.filetype.match, { filename = filename })
			if ok and detected then
				ft = detected
				vim.bo[bufnr].filetype = detected
			end
		else
			vim.bo[bufnr].filetype = ""
		end
		vim.bo[bufnr].modifiable = false
		return ft
	`, &filetype, bufNr, filename, skipHighlight)
	return filetype, err
}

func (e *FileExplorer) positionPreviewWindowAtTop() {
	if e.previewWinID < 0 {
		return
	}
	if err := e.nvim.ExecLua(`
		local win = ...
		if not vim.api.nvim_win_is_valid(win) then
			return
		end
		pcall(vim.api.nvim_win_set_cursor, win, {1, 0})
	`, nil, e.previewWinID); err != nil {
		log.Warnf("[FileExplorer] reset preview scroll failed: %v", err)
	}
}

func (e *FileExplorer) restoreCurrentPaneFocus() {
	if e.currentWinID < 0 {
		return
	}
	if err := e.nvim.SetCurrentWindow(e.currentWinID); err != nil {
		log.Warnf("[FileExplorer] restore current pane focus failed: %v", err)
	}
}

func (e *FileExplorer) requestRedraw() {
	if err := e.nvim.Command("redraw"); err != nil {
		log.Debugf("[FileExplorer] redraw request failed: %v", err)
	}
}

func (e *FileExplorer) ResizePreviewPixels(widthPx, heightPx int) {
	cols := widthPx / previewCellWidthPx
	rows := heightPx / previewCellHeightPx
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	e.previewRows = rows
	if e.previewWinID >= 0 {
		if err := e.nvim.SetWindowSize(e.previewWinID, cols, rows); err != nil {
			log.Warnf("[FileExplorer] resize preview failed: %v", err)
		}
	}
	e.restoreCurrentPaneFocus()
	if e.previewKind == PreviewKindTextFile && e.previewPath != "" {
		entry := DirectoryEntry{Path: e.previewPath, Text: e.previewTitle}
		e.previewGeneration++
		if err := e.refreshPreviewForEntry(entry, e.previewGeneration); err != nil {
			log.Warnf("[FileExplorer] resize preview reload failed: %v", err)
		}
	}
}

func readPreviewLines(filename string, limit int) ([]string, bool, bool, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, false, false, err
	}
	if info.IsDir() {
		return nil, false, false, os.ErrInvalid
	}
	huge := info.Size() > previewHugeFileBytes
	file, err := os.Open(filename)
	if err != nil {
		return nil, false, huge, err
	}
	defer file.Close()

	probe := make([]byte, previewBinaryProbe)
	n, err := file.Read(probe)
	if err != nil && err != io.EOF {
		return nil, false, huge, err
	}
	for _, b := range probe[:n] {
		if b == 0 {
			return nil, true, huge, nil
		}
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, false, huge, err
	}

	reader := bufio.NewReader(file)
	lines := make([]string, 0, limit)
	for len(lines) < limit {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, false, huge, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line != "" || err != io.EOF {
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines, false, huge, nil
}

func stringsToBytes(lines []string) [][]byte {
	result := make([][]byte, len(lines))
	for i, line := range lines {
		result[i] = []byte(line)
	}
	return result
}

func isPreviewImage(filename string) bool {
	return previewImageExtensions[strings.ToLower(path.Ext(filename))]
}

func previewEntryTitle(entry DirectoryEntry) string {
	if strings.TrimSpace(entry.Text) != "" {
		return entry.Text
	}
	return path.Base(entry.Path)
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
	e.stopPreviewTimer()
	e.parent = nil
	e.current = nil
	e.preview = Directory{}
	e.previewKind = PreviewKindNone
	e.previewPath = ""
	e.previewTitle = ""
	e.previewFiletype = ""
	e.previewTextLines = nil
	e.previewTextBufNr = 0
	e.previewTextBufPath = ""
	e.parentWinID = -1
	e.currentWinID = -1
	e.previewWinID = -1
	e.directoriesByPath = map[string]*Directory{}
	e.directoriesByBuf = map[int]*Directory{}
	e.dirtyByBuf = map[int]*DirDraft{}
	e.sourceByID = map[uint64]DirectoryEntry{}
	e.idByPath = map[string]uint64{}
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
		entryPath := path + string(os.PathSeparator) + dir
		return DirectoryEntry{
			ID:        e.idForPath(entryPath),
			Icon:      DIR_ICON,
			IconClass: "",
			Text:      dir,
			IsDir:     true,
			Path:      entryPath,
		}
	})
	files := utils.MapArray(fileStrings, func(file string) DirectoryEntry {
		entryPath := path + string(os.PathSeparator) + file
		return DirectoryEntry{
			ID:        e.idForPath(entryPath),
			Icon:      FILE_ICON,
			IconClass: "",
			Text:      file,
			Path:      entryPath,
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

func (e *FileExplorer) idForPath(entryPath string) uint64 {
	if e.idByPath == nil {
		e.idByPath = map[string]uint64{}
	}
	if id, ok := e.idByPath[entryPath]; ok {
		return id
	}
	id := e.nextID()
	e.idByPath[entryPath] = id
	return id
}

func (e *FileExplorer) setBufferLinesFromEntries(bufNr int, entries []DirectoryEntry) error {
	target := e.entriesToBufferLines(entries)
	existing, err := e.nvim.GetBufferLines(bufNr, 0, -1, false)
	if err != nil {
		return err
	}
	if bufferLinesEqual(existing, target) {
		return nil
	}
	return e.nvim.SetBufferLines(bufNr, 0, -1, false, target)
}

func bufferLinesEqual(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if string(a[i]) != string(b[i]) {
			return false
		}
	}
	return true
}

func (e *FileExplorer) syncPaneBuffers() error {
	if e.preview.BufNr == 0 || e.previewKind == PreviewKindTextFile || e.previewKind == PreviewKindLocalImage {
		return nil
	}
	if err := e.setBufferLinesFromEntries(e.preview.BufNr, e.preview.Entries); err != nil {
		return err
	}
	return nil
}
