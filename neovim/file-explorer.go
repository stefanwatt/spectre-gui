package neovim

import "strings"

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
	ID   int    `json:"id"`
	Icon string `json:"icon"`
	Text string `json:"text"`
}

type FileExplorerDirectory struct {
	WinId   int                          `json:"winId"`
	BufNr   int                          `json:"bufNr"`
	Entries []FileExplorerDirectoryEntry `json:"entries"`
}

type FileExplorer struct {
	Parent  FileExplorerDirectory `json:"parent"`
	Current FileExplorerDirectory `json:"current"`
	Preview FileExplorerPreview   `json:"preview"`
}

func NewFileExplorer() *FileExplorer {
	return &FileExplorer{
		Parent: FileExplorerDirectory{
			BufNr:   -1,
			Entries: []FileExplorerDirectoryEntry{},
		},
		Current: FileExplorerDirectory{
			BufNr:   -1,
			Entries: []FileExplorerDirectoryEntry{},
		},
		Preview: FileExplorerDirectoryPreview{
			Directory: &FileExplorerDirectory{
				BufNr:   -1,
				Entries: []FileExplorerDirectoryEntry{},
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

		entries = append(entries, FileExplorerDirectoryEntry{
			ID:   row,
			Icon: icon,
			Text: text,
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
	if preview, ok := fe.Preview.(FileExplorerDirectoryPreview); ok {
		preview.Directory.Entries = fe.renderFileExplorerDirectory(grid)
	}
}

func (fe *FileExplorer) updateContentPreview(content []ContentRow) {
	if preview, ok := fe.Preview.(FileExplorerContentPreview); ok {
		preview.Content = content
	}
}
