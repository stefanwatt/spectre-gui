package neovim

import (
	"fmt"
	"math"
	"nvim-gui/rendering"
	"nvim-gui/utils"
	"path/filepath"

	"github.com/charmbracelet/log"
)

type FileExplorerPreview interface {
	isFileExplorerPreview()
	GetBufNr() int
	SetBufNr(bufNr int)
	GetWinId() int
	SetWinId(winId int)
}

type FileExplorerDirectoryPreview struct {
	Directory *FileExplorerDirectory `json:"directory"`
}

func (FileExplorerDirectoryPreview) isFileExplorerPreview() {}
func (dp FileExplorerDirectoryPreview) GetBufNr() int {
	return dp.Directory.BufNr
}

func (dp FileExplorerDirectoryPreview) SetBufNr(bufNr int) {
	dp.Directory.BufNr = bufNr
}

func (dp FileExplorerDirectoryPreview) GetWinId() int {
	return dp.Directory.WinId
}

func (dp FileExplorerDirectoryPreview) SetWinId(winId int) {
	dp.Directory.WinId = winId
}

type FileExplorerContentPreview struct {
	Content []rendering.ContentRow `json:"content"`
	BufNr   int                    `json:"bufNr"`
	WinId   int                    `json:"winId"`
}

func (FileExplorerContentPreview) isFileExplorerPreview() {}

func (cp *FileExplorerContentPreview) GetBufNr() int {
	return cp.BufNr
}

func (cp *FileExplorerContentPreview) SetBufNr(bufNr int) {
	cp.BufNr = bufNr
}

func (cp *FileExplorerContentPreview) GetWinId() int {
	return cp.WinId
}

func (cp *FileExplorerContentPreview) SetWinId(winId int) {
	cp.WinId = winId
}

type FileExplorerDirectoryEntry struct {
	ID        int    `json:"id"`
	Icon      string `json:"icon"`
	IconClass string `json:"iconClass"`
	Text      string `json:"text"`
	FilePath  string `json:"filePath"`
	IsDir     bool   `json:"isDir"`
}

type FileExplorerDirectory struct {
	WinId           int                          `json:"winId"`
	BufNr           int                          `json:"bufNr"`
	Entries         []FileExplorerDirectoryEntry `json:"entries"`
	SelectedEntryId int                          `json:"selectedEntryId"`
	CursorCol       int                          `json:"cursorCol"`
}

type FileExplorer struct {
	Parent             FileExplorerDirectory `json:"parent"`
	Current            FileExplorerDirectory `json:"current"`
	Preview            FileExplorerPreview   `json:"preview"`
	CurrentWinMode     string                `json:"currentWinMode"`
	PreviewTargetRows  int                   `json:"-"`
	PreviewTargetCols  int                   `json:"-"`
	DirectoryLineIsDir map[int]map[int]bool  `json:"-"`
}

func NewFileExplorer() *FileExplorer {
	return &FileExplorer{
		Parent:             FileExplorerDirectory{BufNr: -1, WinId: -1, Entries: []FileExplorerDirectoryEntry{}, SelectedEntryId: -1},
		Current:            FileExplorerDirectory{BufNr: -1, WinId: -1, Entries: []FileExplorerDirectoryEntry{}, SelectedEntryId: -1},
		Preview:            FileExplorerDirectoryPreview{Directory: &FileExplorerDirectory{BufNr: -1, WinId: -1, Entries: []FileExplorerDirectoryEntry{}, SelectedEntryId: -1}},
		CurrentWinMode:     "normal",
		DirectoryLineIsDir: make(map[int]map[int]bool),
	}
}

func (fe *FileExplorer) SetPreviewTargetSize(cols, rows int) {
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	fe.PreviewTargetCols = cols
	fe.PreviewTargetRows = rows
}

func (fe *FileExplorer) HasPreviewTargetSize() bool {
	return fe.PreviewTargetCols > 0 && fe.PreviewTargetRows > 0
}

func PreviewGridSizeFromPixels(widthPx, heightPx int) (cols, rows int) {
	const cellWidth = 12
	const cellHeight = 28
	if widthPx < 1 {
		widthPx = 1
	}
	if heightPx < 1 {
		heightPx = 1
	}
	cols = int(math.Floor(float64(widthPx) / float64(cellWidth)))
	rows = int(math.Floor(float64(heightPx) / float64(cellHeight)))
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	return cols, rows
}

func SetMiniFilesWindowOverrides(currentCols, previewCols, currentRows, previewRows int) error {
	if currentCols < 1 {
		currentCols = 1
	}
	if previewCols < 1 {
		previewCols = 1
	}
	if currentRows < 1 {
		currentRows = 1
	}
	if previewRows < 1 {
		previewRows = 1
	}
	lua := `
local cur_w, prev_w, cur_h, prev_h = ...
local ok, mf = pcall(require, "mini.files")
if not ok or mf == nil then
  return false
end
if type(mf.set_window_overrides) == "function" then
  mf.set_window_overrides(cur_w, prev_w, cur_h, prev_h)
  return true
end
if type(mf.set_window_width_overrides) == "function" then
  mf.set_window_width_overrides(cur_w, prev_w)
  return true
end
if type(_G.MiniFiles) == "table" then
  if type(_G.MiniFiles.set_window_overrides) == "function" then
    _G.MiniFiles.set_window_overrides(cur_w, prev_w, cur_h, prev_h)
    return true
  end
  if type(_G.MiniFiles.set_window_width_overrides) == "function" then
    _G.MiniFiles.set_window_width_overrides(cur_w, prev_w)
    return true
  end
end
return false
`
	var applied bool
	if err := NvimClient.ExecLua(lua, &applied, currentCols, previewCols, currentRows, previewRows); err != nil {
		return err
	}
	if !applied {
		return fmt.Errorf("mini.files window override API not available")
	}
	return nil
}

func EnsureMiniFilesPatched() error {
	lua := `
if package.loaded["mini.files"] ~= nil then
  package.loaded["mini.files"] = nil
end
if type(_G.MiniFiles) == "table" then
  _G.MiniFiles = nil
end
return true
`
	var ok bool
	if err := NvimClient.ExecLua(lua, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("failed to reset mini.files module before patch")
	}
	patchPath := filepath.Join("/home/stefan/Projects/mini.files/lua", "mini/files.lua")
	loadLua := `
local path = ...
if vim and vim.fn and vim.fn.filereadable(path) == 0 then
  return false
end
local chunk, load_err = loadfile(path)
if not chunk then
  error(load_err)
end
local module = chunk()
package.loaded["mini.files"] = module
_G.MiniFiles = module
return true
`
	if err := NvimClient.ExecLua(loadLua, &ok, patchPath); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("failed to load patched mini.files from %s", patchPath)
	}
	return nil
}

func GetMiniFilesDirectoryLineMap(bufNr int) (map[int]bool, error) {
	lineMap := make(map[int]bool)
	if bufNr < 1 || NvimClient == nil {
		return lineMap, nil
	}

	lua := `
local buf_id = ...
if not vim.api.nvim_buf_is_valid(buf_id) then
  return {}
end

local getter = nil
local ok, mf = pcall(require, "mini.files")
if ok and type(mf) == "table" and type(mf.get_fs_entry) == "function" then
  getter = mf.get_fs_entry
elseif type(_G.MiniFiles) == "table" and type(_G.MiniFiles.get_fs_entry) == "function" then
  getter = _G.MiniFiles.get_fs_entry
end

if getter == nil then
  return {}
end

local line_count = vim.api.nvim_buf_line_count(buf_id)
local directories = {}
for line = 1, line_count do
  local ok_entry, entry = pcall(getter, buf_id, line)
  if ok_entry and type(entry) == "table" and entry.fs_type == "directory" then
    table.insert(directories, line)
  end
end

return directories
`

	var directoryLines []interface{}
	if err := NvimClient.ExecLua(lua, &directoryLines, bufNr); err != nil {
		return lineMap, err
	}
	for _, rawLine := range directoryLines {
		line := utils.ReflectToInt(rawLine)
		if line > 0 {
			lineMap[line] = true
		}
	}
	return lineMap, nil
}

func (fe *FileExplorer) RefreshDirectoryLineMap(bufNr int) {
	if bufNr < 1 {
		return
	}
	lineMap, err := GetMiniFilesDirectoryLineMap(bufNr)
	if err != nil {
		log.Debug(fmt.Sprintf("[minifiles] failed to fetch fs entry types for buf=%d: %v", bufNr, err))
		return
	}
	if fe.DirectoryLineIsDir == nil {
		fe.DirectoryLineIsDir = make(map[int]map[int]bool)
	}
	fe.DirectoryLineIsDir[bufNr] = lineMap
}
