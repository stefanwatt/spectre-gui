package neovim

import (
	"fmt"
	"nvim-gui/utils"
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

func (cp FileExplorerContentPreview) GetBufNr() int {
	return cp.BufNr
}

func (cp FileExplorerContentPreview) SetBufNr(bufNr int) {
	cp.BufNr = bufNr
}

func (cp FileExplorerContentPreview) GetWinId() int {
	return cp.WinId
}

func (cp FileExplorerContentPreview) SetWinId(winId int) {
	cp.WinId = winId
}

type FileExplorerDirectoryEntry struct {
	ID       int    `json:"id"`
	Icon     string `json:"icon"`
	Text     string `json:"text"`
	FilePath string `json:"filePath"`
	IsDir    bool   `json:"isDir"`
}

type FileExplorerDirectory struct {
	WinId           int                          `json:"winId"`
	BufNr           int                          `json:"bufNr"`
	Entries         []FileExplorerDirectoryEntry `json:"entries"`
	SelectedEntryId int                          `json:"selectedEntryId"`
}

type FileExplorer struct {
	Parent  FileExplorerDirectory `json:"parent"`
	Current FileExplorerDirectory `json:"current"`
	Preview FileExplorerPreview   `json:"preview"`
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
	}
}

func (fe *FileExplorer) renderFileExplorerDirectory(grid *Grid) []FileExplorerDirectoryEntry {
	entries := make([]FileExplorerDirectoryEntry, 0, grid.Height)
	for row := 0; row < grid.Height; row++ {
		rowCells := grid.Cells[row]
		if len(rowCells) == 0 {
			continue
		}

		var fullText strings.Builder
		for _, cell := range rowCells {
			fullText.WriteString(cell.Char)
		}

		line := strings.TrimSpace(fullText.String())
		if line == "" {
			continue
		}

		// Skip border rows (top: ┌...┐, bottom: └...┘)
		if strings.HasPrefix(line, "┌") || strings.HasPrefix(line, "└") {
			continue
		}

		// Strip │ borders from content rows
		if strings.HasPrefix(line, "│") && strings.HasSuffix(line, "│") {
			line = strings.TrimPrefix(line, "│")
			line = strings.TrimSuffix(line, "│")
			line = strings.TrimSpace(line)
		}

		if line == "" {
			continue
		}

		// Extract icon: first rune if non-ASCII (e.g., devicon)
		var icon string
		text := line
		runes := []rune(line)
		if len(runes) > 1 && runes[0] > 127 {
			icon = string(runes[0])
			text = strings.TrimSpace(string(runes[1:]))
		}

		if icon == "" && text == "" {
			continue
		}

		isDir := strings.HasSuffix(text, "/")

		entries = append(entries, FileExplorerDirectoryEntry{
			ID:    row,
			Icon:  icon,
			Text:  text,
			IsDir: isDir,
		})
	}

	return entries
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

	fe.Preview = FileExplorerContentPreview{
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
