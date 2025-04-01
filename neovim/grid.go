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
}

func (g *Grid) toString() string {
	result := ""
	for _, row := range g.Cells {
		for _, cell := range row {
			result += cell.Char
		}
		result += "\n"
	}
	return result
}

func (g *Grid) toHexGrid() *Grid {
	newGrid := &Grid{
		ID:     g.ID,
		Width:  g.Width,
		Height: g.Height,
	}
	newGrid.Cursor.Row = g.Cursor.Row
	newGrid.Cursor.Col = g.Cursor.Col
	newGrid.Cells = make([][]*Cell, len(g.Cells))
	for i, row := range g.Cells {
		newGrid.Cells[i] = make([]*Cell, len(row))
		for j, cell := range row {
			newCell := &Cell{}
			var hexChar string
			for _, r := range cell.Char {
				hexChar += fmt.Sprintf("%x", r)
			}
			newCell.Char = hexChar
			newCell.Highlight = cell.Highlight
			newGrid.Cells[i][j] = newCell
		}
	}
	return newGrid
}

type Cell struct {
	Char      string `json:"char"`
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
