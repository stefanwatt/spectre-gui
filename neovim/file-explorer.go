package neovim

import (
	"fmt"
	"math"
	"nvim-gui/utils"
	"path/filepath"
	"strings"
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
	Content []ContentRow `json:"content"`
	BufNr   int          `json:"bufNr"`
	WinId   int          `json:"winId"`
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
}

type FileExplorer struct {
	Parent            FileExplorerDirectory `json:"parent"`
	Current           FileExplorerDirectory `json:"current"`
	Preview           FileExplorerPreview   `json:"preview"`
	PreviewTargetRows int                   `json:"-"`
	PreviewTargetCols int                   `json:"-"`
}

func NewFileExplorer() *FileExplorer {
	return &FileExplorer{
		Parent: FileExplorerDirectory{
			BufNr:           -1,
			WinId:           -1,
			Entries:         []FileExplorerDirectoryEntry{},
			SelectedEntryId: -1,
		},
		Current: FileExplorerDirectory{
			BufNr:           -1,
			WinId:           -1,
			Entries:         []FileExplorerDirectoryEntry{},
			SelectedEntryId: -1,
		},
		Preview: FileExplorerDirectoryPreview{
			Directory: &FileExplorerDirectory{
				BufNr:           -1,
				WinId:           -1,
				Entries:         []FileExplorerDirectoryEntry{},
				SelectedEntryId: -1,
			},
		},
		PreviewTargetRows: 0,
		PreviewTargetCols: 0,
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
	if err := NvimInstance.ExecLua(lua, &applied, currentCols, previewCols, currentRows, previewRows); err != nil {
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
	if err := NvimInstance.ExecLua(lua, &ok); err != nil {
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
	if err := NvimInstance.ExecLua(loadLua, &ok, patchPath); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("failed to load patched mini.files from %s", patchPath)
	}
	return nil
}

func (fe *FileExplorer) renderFileExplorerDirectory(grid *Grid) []FileExplorerDirectoryEntry {
	entries := make([]FileExplorerDirectoryEntry, 0, grid.Height)
	for row := 0; row < grid.Height; row++ {
		rowCells := grid.Cells[row]
		if len(rowCells) == 0 {
			continue
		}

		firstNonEmpty := -1
		lastNonEmpty := -1
		for i, cell := range rowCells {
			if strings.TrimSpace(cell.Char) != "" {
				firstNonEmpty = i
				break
			}
		}
		if firstNonEmpty == -1 {
			continue
		}
		for i := len(rowCells) - 1; i >= 0; i-- {
			if strings.TrimSpace(rowCells[i].Char) != "" {
				lastNonEmpty = i
				break
			}
		}
		if lastNonEmpty == -1 {
			continue
		}

		leftChar := rowCells[firstNonEmpty].Char
		rightChar := rowCells[lastNonEmpty].Char

		// Skip border rows (top: ┌...┐, bottom: └...┘)
		if leftChar == "┌" || leftChar == "└" {
			continue
		}

		contentStart := firstNonEmpty
		contentEnd := lastNonEmpty
		if leftChar == "│" && rightChar == "│" && contentEnd-contentStart >= 2 {
			contentStart++
			contentEnd--
		}
		if contentStart > contentEnd {
			continue
		}
		contentCells := rowCells[contentStart : contentEnd+1]

		line := strings.TrimSpace(cellsToString(contentCells))
		if line == "" {
			continue
		}

		var icon string
		var iconClass string
		text := line
		iconCellIndex := -1
		for i, cell := range contentCells {
			if strings.TrimSpace(cell.Char) != "" {
				iconCellIndex = i
				break
			}
		}
		if iconCellIndex >= 0 {
			iconRunes := []rune(contentCells[iconCellIndex].Char)
			if len(iconRunes) > 0 && iconRunes[0] > 127 {
				icon = contentCells[iconCellIndex].Char
				iconClass = contentCells[iconCellIndex].ClassesToString()
				text = strings.TrimSpace(cellsToString(contentCells[iconCellIndex+1:]))
			}
		}

		if icon == "" && text == "" {
			continue
		}

		isDir := strings.HasSuffix(text, "/")

		entries = append(entries, FileExplorerDirectoryEntry{
			ID:        row,
			Icon:      icon,
			IconClass: iconClass,
			Text:      text,
			IsDir:     isDir,
		})
	}

	return entries
}

func cellsToString(cells []*Cell) string {
	var builder strings.Builder
	for _, cell := range cells {
		if cell == nil {
			continue
		}
		builder.WriteString(cell.Char)
	}
	return builder.String()
}

func (fe *FileExplorer) updateParent(grid *Grid) {
	fe.Parent.Entries = fe.renderFileExplorerDirectory(grid)
}

func (fe *FileExplorer) updateCurrent(grid *Grid) {
	fe.Current.Entries = fe.renderFileExplorerDirectory(grid)
}

func (fe *FileExplorer) updateDirPreview(grid *Grid) {
	winId := fe.Preview.GetWinId()
	bufNr := fe.Preview.GetBufNr()
	fe.Preview = FileExplorerDirectoryPreview{
		Directory: &FileExplorerDirectory{
			WinId:           winId,
			BufNr:           bufNr,
			Entries:         fe.renderFileExplorerDirectory(grid),
			SelectedEntryId: -1,
		},
	}
}

func (fe *FileExplorer) updateContentPreview(content []ContentRow) {
	winId := fe.Preview.GetWinId()
	bufNr := fe.Preview.GetBufNr()
	if len(content) > 2 {
		content = content[1 : len(content)-1]
	}
	for i := range content {
		if len(content[i].Tokens) > 2 {
			content[i].Tokens = content[i].Tokens[1:]
			if len(content[i].Tokens) > 0 {
				content[i].Tokens = content[i].Tokens[:len(content[i].Tokens)-1]
			}
		}
	}

	fe.Preview = &FileExplorerContentPreview{
		Content: content,
		BufNr:   bufNr,
		WinId:   winId,
	}
}

func (fe *FileExplorer) updateCursor(winId, row, col int) (*FileExplorerDirectoryEntry, error) {
	// Helper to find the entry ID for a given row in a directory
	findEntryForRow := func(entries []FileExplorerDirectoryEntry, targetRow int) (*FileExplorerDirectoryEntry, error) {
		// Find the entry whose ID matches the grid row
		for _, entry := range entries {
			if entry.ID == targetRow {
				utils.Log(fmt.Sprintf("[fileexplorer] found entry with id=%d", winId))
				return &entry, nil
			}
		}
		utils.Log(fmt.Sprintf("[fileexplorer] could not find entry with id=%d", winId))
		return nil, fmt.Errorf("[fileexplorer] could not find entry with id=%d", winId)
	}

	if fe.Parent.WinId == winId {
		utils.Log("[fileexplorer] updating cursor for parent")
		entry, err := findEntryForRow(fe.Parent.Entries, row)
		if err != nil {
			return nil, err
		}
		fe.Parent.SelectedEntryId = entry.ID
		return entry, nil
	} else if fe.Current.WinId == winId {
		utils.Log("[fileexplorer] updating cursor for current")
		entry, err := findEntryForRow(fe.Current.Entries, row)
		if err != nil {
			return nil, err
		}
		fe.Current.SelectedEntryId = entry.ID
		return entry, nil
	} else if fe.Preview.GetWinId() == winId {
		utils.Log("[fileexplorer] updating cursor for preview")
		if dirPreview, ok := fe.Preview.(FileExplorerDirectoryPreview); ok {
			entry, err := findEntryForRow(dirPreview.Directory.Entries, row)
			if err != nil {
				return nil, err
			}
			dirPreview.Directory.SelectedEntryId = entry.ID
			return entry, nil
		}
	}
	return nil, fmt.Errorf("didnt find window or entry for cursor")
}
