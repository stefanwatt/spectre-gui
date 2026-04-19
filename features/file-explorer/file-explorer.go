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

	"github.com/charmbracelet/log"
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
	err = e.nvim.SetBufferToWindow(parentWin, e.parent.BufNr)
	if err != nil {
		return err
	}
	var currentWin int
	err = e.nvim.OpenSplitRight(&currentWin, e.current.BufNr)
	if err != nil {
		return err
	}
	var previewWin int
	err = e.nvim.OpenSplitRight(&previewWin, e.preview.BufNr)
	if err != nil {
		return err
	}

	selectedParentEntry, _ := utils.Find(parentEntries, func(entry DirectoryEntry) bool {
		return strings.Contains(filepath, entry.Path)
	})

	e.parent = Directory{
		WinID:           parentWin,
		BufNr:           e.parent.BufNr,
		Entries:         parentEntries,
		SelectedEntryId: selectedParentEntry.ID,
	}

	selectedCurrentEntry, _ := utils.Find(currentEntries, func(entry DirectoryEntry) bool {
		return filepath == entry.Path
	})
	e.current = Directory{
		WinID:           currentWin,
		BufNr:           e.current.BufNr,
		Entries:         currentEntries,
		SelectedEntryId: selectedCurrentEntry.ID,
	}
	e.preview = Directory{
		WinID:   previewWin,
		BufNr:   e.preview.BufNr,
		Entries: []DirectoryEntry{},
	}
	e.active = true
	e.Dirty = true
	err = e.nvim.CreateBufferKeymap(
		e.current.BufNr,
		"n",
		"q",
		func(channelID int) string {
			return fmt.Sprintf(":lua vim.rpcnotify(%d, 'FileExplorerClose', {})<CR>", channelID)
		},
	)
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Info("successfully set up keymap for closing fileexplorer")
	}

	e.nvim.SetCurrentWindow(e.current.WinID)

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
