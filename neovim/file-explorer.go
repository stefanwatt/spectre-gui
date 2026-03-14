package neovim

import "strings"

type FileExplorerPreview interface {
	isFileExplorerPreview()
	GetBufNr() int
	SetBufNr(bufNr int)
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

type FileExplorerContentPreview struct {
	Content []ContentRow `json:"content"`
	BufNr   int          `json:"bufNr"`
}

func (FileExplorerContentPreview) isFileExplorerPreview() {}

func (cp FileExplorerContentPreview) GetBufNr() int {
	return cp.BufNr
}

func (cp FileExplorerContentPreview) SetBufNr(bufNr int) {
	cp.BufNr = bufNr
}

type FileExplorerDirectoryEntry struct {
	ID   int    `json:"id"`
	Icon string `json:"icon"`
	Text string `json:"text"`
}

type FileExplorerDirectory struct {
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

		line := fullText.String()
		if line == "" {
			continue
		}
		var icon string
		var text string
		if len(line) > 0 && line[0] == '/' {
			iconEnd := strings.Index(line[1:], "/")
			if iconEnd != -1 {
				icon = line[1 : iconEnd+1] // Extract icon between slashes
				remainder := line[iconEnd+2:]
				remainder = strings.TrimLeft(remainder, " ")
				if len(remainder) > 0 && remainder[0] == '/' {
					remainder = remainder[1:]
				}

				text = strings.TrimSpace(remainder)
			} else {
				text = strings.TrimSpace(line)
			}
		} else {
			text = strings.TrimSpace(line)
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
