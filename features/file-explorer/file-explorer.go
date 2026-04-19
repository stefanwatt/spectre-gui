package fileexplorer

import (
	"fmt"
	"os"
	path "path/filepath"
	"sort"
	"strings"
	"sync/atomic"

	"nvim-gui/core/ports"
	"nvim-gui/utils"
)

type FileExplorer struct {
	active     bool
	Dirty      bool
	idCounter  uint64
	nvim       ports.NvimClient
	parent     Directory
	current    Directory
	preview    Directory
	parentBuf  int
	parentWin  int
	currentBuf int
	currentWin int
	previewBuf int
	previewWin int
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
	return e.currentBuf
}

func (e *FileExplorer) GetCurrentWin() int {
	return e.currentWin
}
func (e *FileExplorer) GetParentBuf() int {
	return e.parentBuf
}
func (e *FileExplorer) GetParentWin() int {
	return e.parentWin
}
func (e *FileExplorer) GetPreviewBuf() int {
	return e.previewBuf
}
func (e *FileExplorer) GetPreviewWin() int {
	return e.previewWin
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
	e.parentBuf, err = e.nvim.CreateBuffer(true, false)
	if err != nil {
		return err
	}
	e.currentBuf, err = e.nvim.CreateBuffer(true, false)
	if err != nil {
		return err
	}
	e.previewBuf, err = e.nvim.CreateBuffer(true, false)
	if err != nil {
		return err
	}
	currentDir := path.Dir(filepath)
	parentDir := path.Dir(currentDir)
	if parentDir == "." {
		return fmt.Errorf("[FileExplorer] parent dir doesnt exist")
	}
	parentEntries, err := e.mapDirectoryEntries(parentDir)
	if err != nil {
		return err
	}
	currentEntries, err := e.mapDirectoryEntries(currentDir)
	if err != nil {
		return err
	}
	err = e.nvim.Command("tab new")
	if err != nil {
		return err
	}
	parentWin, err := e.nvim.CurrentWindow()
	if err != nil {
		return err
	}
	err = e.nvim.SetBufferToWindow(parentWin, e.parentBuf)
	if err != nil {
		return err
	}
	var currentWin int
	err = e.nvim.OpenSplitRight(&currentWin, e.currentBuf)
	if err != nil {
		return err
	}
	var previewWin int
	err = e.nvim.OpenSplitRight(&previewWin, e.previewBuf)
	if err != nil {
		return err
	}

	selectedParentEntry, _ := utils.Find(parentEntries, func(entry DirectoryEntry) bool {
		return strings.Contains(filepath, entry.Path)
	})

	e.parent = Directory{
		WinID:           parentWin,
		BufNr:           e.parentBuf,
		Entries:         parentEntries,
		SelectedEntryId: selectedParentEntry.ID,
	}

	selectedCurrentEntry, _ := utils.Find(currentEntries, func(entry DirectoryEntry) bool {
		return filepath == entry.Path
	})
	e.current = Directory{
		WinID:           currentWin,
		BufNr:           e.currentBuf,
		Entries:         currentEntries,
		SelectedEntryId: selectedCurrentEntry.ID,
	}
	e.preview = Directory{
		WinID:   previewWin,
		BufNr:   e.previewBuf,
		Entries: []DirectoryEntry{},
	}
	e.active = true
	e.Dirty = true

	return err
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

func (e *FileExplorer) GoIn() error {
	var err error
	e.Dirty = true
	return err
}

func (e *FileExplorer) GoOut() error {
	var err error
	e.Dirty = true
	return err
}
