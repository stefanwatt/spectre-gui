package fileexplorer

import (
	"fmt"
	"os"
	path "path/filepath"
	"sort"

	"nvim-gui/core/ports"
	"nvim-gui/utils"

	"github.com/charmbracelet/log"
)

type FileExplorer struct {
	registry   *Registry
	nvim       ports.NvimClient
	parentBuf  int
	parentWin  int
	currentBuf int
	currentWin int
	previewBuf int
	previewWin int
}

func New(nvim ports.NvimClient, registry *Registry) *FileExplorer {
	return &FileExplorer{
		registry: registry,
		nvim:     nvim,
	}
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
	parentLines, err := e.mapDirectoryLines(parentDir)
	if err != nil {
		return err
	}
	err = e.nvim.SetBufferLines(e.parentBuf, 0, 1, true, parentLines)
	if err != nil {
		return err
	}
	currentLines, err := e.mapDirectoryLines(currentDir)
	err = e.nvim.SetBufferLines(e.currentBuf, 0, 1, true, currentLines)
	if err != nil {
		return err
	}
	err = e.nvim.SetBufferLines(e.previewBuf, 0, 1, true, [][]byte{[]byte("preview")})
	if err != nil {
		return err
	}
	err = e.nvim.Command("tab new")
	if err != nil {
		return err
	}
	e.parentWin, err = e.nvim.CurrentWindow()
	if err != nil {
		return err
	}
	err = e.nvim.SetBufferToWindow(e.parentWin, e.parentBuf)
	if err != nil {
		return err
	}
	err = e.nvim.OpenSplitRight(&e.currentWin, e.currentBuf)
	if err != nil {
		return err
	}
	err = e.nvim.OpenSplitRight(&e.previewWin, e.previewBuf)
	if err != nil {
		return err
	}
	log.Infof("[FileExplorer] parentWin=%d currentWin=%d previewWin=%d", e.parentWin, e.currentWin, e.previewWin)
	err = e.nvim.SetWindowOption(e.parentWin, "relativenumber", false)
	if err != nil {
		return err
	}
	err = e.nvim.SetWindowOption(e.currentWin, "relativenumber", false)
	if err != nil {
		return err
	}
	err = e.nvim.SetWindowOption(e.parentWin, "number", false)
	if err != nil {
		return err
	}
	err = e.nvim.SetWindowOption(e.currentWin, "number", false)
	if err != nil {
		return err
	}
	err = e.nvim.SetWindowOption(e.previewWin, "relativenumber", false)
	return err
	// registry := GetFileExplorerRegistry()
	// registry.AssignCurrentPane(int(parentWin), int(parentBuf))
	// registry.AssignCurrentPane(int(currentWin), int(currentBuf))
	// registry.AssignCurrentPane(int(previewWin), int(previewBuf))
	// registry.SetActive(true)
}

func (e *FileExplorer) mapDirectoryLines(path string) ([][]byte, error) {
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
	dirs := utils.MapArray(dirStrings, func(dir string) []byte { return []byte(dir) })
	files := utils.MapArray(fileStrings, func(file string) []byte { return []byte(file) })

	return append(dirs, files...), nil
}

func (e *FileExplorer) GoIn() error {
	var err error
	return err
}

func (e *FileExplorer) GoOut() error {
	var err error
	return err
}
