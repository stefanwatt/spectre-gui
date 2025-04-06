package neovim

import (
	"fmt"
	"strings"
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
	Char      string
	Highlight int
	Dirty     bool
	Classes   map[string]bool
}

func (c *Cell) ClassesToString() string {
	var builder strings.Builder
	for class, _ := range c.Classes {
		builder.WriteString(class)
		builder.WriteString(" ")
	}
	return strings.TrimRight(builder.String(), " ")
}

func (c *Cell) Equals(other *Cell) bool {
	if c == nil || other == nil {
		return c == other
	}
	return c.Char == other.Char && c.ClassesToString() == other.ClassesToString() && c.Highlight == other.Highlight
}
