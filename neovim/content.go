package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"regexp"
	"strings"
)

type Token struct {
	Text      string `json:"text" msgpack:"text"`
	Classes   string `json:"classes" msgpack:"classes"`
	Highlight int    `json:"highlight" msgpack:"hl_group"`
}

type MarkdownOpts struct {
	QuoteLevel int `json:"quoteLevel"`
}

func NewMarkdownOpts() *MarkdownOpts {
	return &MarkdownOpts{QuoteLevel: 0}
}

type ContentRow struct {
	Index        int           `json:"index" msgpack:"row"` // basically the line number
	Tokens       []*Token      `json:"tokens" msgpack:"tokens"`
	MarkdownOpts *MarkdownOpts `json:"markdownOpts"`
}

func (cr *ContentRow) ToString() string {
	var sb strings.Builder
	for _, token := range cr.Tokens {
		sb.WriteString(token.Text)
	}
	return sb.String()
}

func fromCell(cell *Cell) *Token {
	return &Token{
		Text:      cell.Char,
		Classes:   cell.ClassesToString(),
		Highlight: cell.Highlight,
	}
}

func (t *Token) toCell() *Cell {
	classes := make(map[string]bool)
	for _, class := range strings.Split(t.Classes, " ") {
		classes[class] = true
	}
	return &Cell{
		Char:      t.Text,
		Classes:   classes,
		Dirty:     false,
		Highlight: t.Highlight,
	}
}

func MapTokens(cells []*Cell) []*Token {
	return utils.MapArray(cells, func(cell *Cell) *Token {
		return fromCell(cell)
	})
}

func (s *Screen) optimizeGrid(grid *Grid, filetype string) []ContentRow {
	contentRows := make([]ContentRow, grid.Height)
	for row := 0; row < grid.Height; row++ {
		// TODO: handle folds — folded lines cause line numbers to jump,
		// topline+row won't be correct when folds are present.
		lineNumber := grid.TopLine + row + 1 // 1-indexed buffer line
		rowCells := grid.Cells[row]

		if grid.DirtyRows[row] {
			cells := s.optimizeRow(rowCells, row, grid.Cursor)
			tokens := MapTokens(cells)
			contentRows[row].Tokens = tokens
			contentRows[row].Index = lineNumber
			grid.OptimizedRows[row] = cells
			grid.DirtyRows[row] = false

			if filetype == "markdown" {
				markdownOpts := getMarkdownOpts(contentRows[row])
				contentRows[row].MarkdownOpts = markdownOpts
				grid.MarkdownOpts[row] = markdownOpts
			}
		} else {
			contentRows[row].Tokens = MapTokens(grid.OptimizedRows[row])
			contentRows[row].Index = lineNumber
			if opts, exists := grid.MarkdownOpts[row]; exists {
				contentRows[row].MarkdownOpts = opts
			} else {
				contentRows[row].MarkdownOpts = getMarkdownOpts(contentRows[row])
			}
		}
	}
	return contentRows
}

func getMarkdownOpts(contentRow ContentRow) *MarkdownOpts {
	text := contentRow.ToString()
	re := regexp.MustCompile(`^(\s*>)+`)
	match := re.FindString(text)
	count := 0
	for _, r := range match {
		if r == '>' {
			count++
		}
	}
	return &MarkdownOpts{QuoteLevel: count}
}

func (s *Screen) optimizeRow(rowCells []*Cell, currentRow int, cursor struct {
	Row int
	Col int
}) []*Cell {
	optimizedRow := make([]*Cell, 0)

	if len(rowCells) == 0 {
		return optimizedRow
	}

	var currentToken *Cell
	lastHl := 0

	for col := 0; col < len(rowCells); col++ {
		cell := rowCells[col]
		isCursor := cursor.Row == currentRow && cursor.Col == col

		if isCursor {
			if currentToken != nil {
				optimizedRow = append(optimizedRow, currentToken)
				currentToken = nil
			}
			classes := make(map[string]bool)
			for class, _ := range cell.Classes {
				classes[class] = true
			}
			classes["cursor"] = true
			cursorCell := &Cell{
				Char:      cell.Char,
				Highlight: cell.Highlight,
				Classes:   classes,
			}
			optimizedRow = append(optimizedRow, cursorCell)
			lastHl = cell.Highlight
			continue
		}

		if cell.Highlight != lastHl || currentToken == nil {
			if currentToken != nil {
				optimizedRow = append(optimizedRow, currentToken)
			}
			currentToken = &Cell{
				Char:      cell.Char,
				Highlight: cell.Highlight,
				Classes:   cell.Classes,
			}
			lastHl = cell.Highlight
		} else {
			currentToken.Char += cell.Char
		}
	}

	if currentToken != nil {
		optimizedRow = append(optimizedRow, currentToken)
	}

	// Trim trailing whitespace
	if len(optimizedRow) > 0 {
		lastToken := optimizedRow[len(optimizedRow)-1]
		lastToken.Char = strings.TrimRight(lastToken.Char, " ")
	}

	return optimizedRow
}

func sanitize(s string) string {
	s = strings.Replace(s, " ", `&nbsp;`, -1)
	s = strings.Replace(s, "\t", `&nbsp;`, -1)
	s = strings.Replace(s, "<", `&lt;`, -1)
	s = strings.Replace(s, ">", `&gt;`, -1)
	return s
}

func (g *Grid) toHex() []ContentRow {
	newContentRows := make([]ContentRow, len(g.Cells))
	for i, row := range g.Cells {
		newContentRows[i].Tokens = make([]*Token, len(row))
		for j, cell := range row {
			newToken := &Token{}
			var hexChar string
			for _, r := range cell.Char {
				hexChar += fmt.Sprintf("%x", r)
			}
			newToken.Text = hexChar
			newToken.Highlight = cell.Highlight
			newToken.Classes = cell.ClassesToString()
			newContentRows[i].Tokens[j] = newToken
		}
	}
	return newContentRows
}
