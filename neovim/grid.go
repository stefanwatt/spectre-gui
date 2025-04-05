package neovim

import "fmt"

type Grid struct {
	ID     int
	Width  int
	Height int
	Cells  [][]*Cell
	Cursor struct {
		Row int
		Col int
	}
	DirtyRows     []bool
	OptimizedRows [][]*Cell
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

func (g *Grid) toHex() []ContentRow {
	newContentRows := make([]ContentRow, len(g.Cells))
	for i, row := range g.Cells {
		newContentRows[i].Tokens = make([]*Cell, len(row))
		for j, cell := range row {
			newCell := &Cell{}
			var hexChar string
			for _, r := range cell.Char {
				hexChar += fmt.Sprintf("%x", r)
			}
			newCell.Char = hexChar
			newCell.Highlight = cell.Highlight
			newContentRows[i].Tokens[j] = newCell
		}
	}
	return newContentRows
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

type Cell struct {
	Char      string `json:"text"`
	Highlight int    `json:"highlight"`
	Dirty     bool
	Classes   string `json:"classes"`
}

func (c *Cell) Equals(other *Cell) bool {
	if c == nil || other == nil {
		return c == other
	}
	return c.Char == other.Char && c.Classes == other.Classes && c.Highlight == other.Highlight
}
