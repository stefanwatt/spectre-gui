package neovim

import (
	"fmt"
	"nvim-gui/rendering"
)

type Grid struct {
	ID     int
	Width  int
	Height int
	Cells  [][]*Cell
	Cursor struct {
		Row int
		Col int
	}
	TopLine       int // first visible buffer line (0-indexed), from win_viewport
	DirtyRows     []bool
	OptimizedRows [][]*Cell
	CachedTokens  [][]*rendering.Token // cached token slices for non-dirty rows
	MarkdownOpts  map[int]*rendering.MarkdownOpts
}

func NewGrid(rows, cols int) *Grid {
	content := make([][]*Cell, rows)
	for i := range content {
		content[i] = make([]*Cell, cols)
		for j := range content[i] {
			content[i][j] = &Cell{
				Char:      " ",
				Highlight: 0,
			}
		}
	}
	return &Grid{
		ID:           1,
		Width:        cols,
		Height:       rows,
		Cells:        content,
		MarkdownOpts: make(map[int]*rendering.MarkdownOpts),
	}
}

func (g *Grid) toString() string {
	result := ""
	if g == nil {
		return ""
	}
	for _, row := range g.Cells {
		for _, cell := range row {
			result += cell.Char
		}
		result += "\n"
	}
	return result
}

func toHexCells(cells [][]*Cell) [][]Cell {
	newCells := make([][]Cell, len(cells))
	for i, row := range cells {
		newCells[i] = make([]Cell, len(row))
		for j, cell := range row {
			var hexChar string
			for _, r := range cell.Char {
				hexChar += fmt.Sprintf("%x", r)
			}
			newCells[i][j] = Cell{
				Char:      hexChar,
				Highlight: cell.Highlight,
			}
		}
	}
	return newCells
}

func (g *Grid) Resize(width, height int) {
	g.Height = height
	g.Width = width

	newCells := make([][]*Cell, height)
	for i := range newCells {
		newCells[i] = make([]*Cell, width)
		for j := range newCells[i] {
			// Copy existing cell if available
			if i < len(g.Cells) && j < len(g.Cells[i]) && g.Cells[i][j] != nil {
				newCells[i][j] = &Cell{
					Char:      g.Cells[i][j].Char,
					Highlight: g.Cells[i][j].Highlight,
					Classes:   g.Cells[i][j].Classes,
				}
			} else {
				newCells[i][j] = &Cell{
					Char:      " ",
					Highlight: 0,
				}
			}
		}
	}

	g.Cells = newCells

	g.DirtyRows = make([]bool, height)
	g.OptimizedRows = make([][]*Cell, height)
	g.CachedTokens = make([][]*rendering.Token, height)
	for i := range g.DirtyRows {
		g.DirtyRows[i] = true
	}
}
