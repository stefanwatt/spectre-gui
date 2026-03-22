package neovim

import (
	"fmt"
	"sort"
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
	TopLine       int // first visible buffer line (0-indexed), from win_viewport
	DirtyRows     []bool
	OptimizedRows [][]*Cell
	CachedTokens  [][]*Token // cached token slices for non-dirty rows
	MarkdownOpts  map[int]*MarkdownOpts
}

func NewGrid(rows int, cols int) *Grid {
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
		MarkdownOpts: make(map[int]*MarkdownOpts),
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

type Cell struct {
	Char      string
	Highlight int
	Dirty     bool
	Classes   map[string]bool
}

func (c *Cell) ClassesToString() string {
	var builder strings.Builder
	var classes []string
	for class, _ := range c.Classes {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	for _, class := range classes {
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
