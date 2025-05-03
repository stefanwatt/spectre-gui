package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"regexp"
	"strconv"
	"strings"
)

type Token struct {
	Text      string `json:"text" msgpack:"text"`
	Classes   string `json:"classes" msgpack:"classes"`
	Highlight int    `json:"highlight" msgpack:"hl_group"`
}

type ContentRow struct {
	Index  int      `json:"index" msgpack:"row"` // basically the line number
	Tokens []*Token `json:"tokens" msgpack:"tokens"`
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
		Highlight: t.Highlight, // check if we really need this
	}
}

func MapTokens(cells []*Cell) []*Token {
	return utils.MapArray(cells, func(cell *Cell) *Token {
		return fromCell(cell)
	})
}

func (s *Screen) optimizeGrid(grid *Grid) []ContentRow {
	contentRows := make([]ContentRow, grid.Height)
	for row := 0; row < grid.Height; row++ {
		rowCells, lineNumber := s.trimGutter(grid.Cells[row])
		if lineNumber < 1 {
			rowCells = grid.Cells[row]
			lineNumber = row
		}
		if grid.DirtyRows[row] {
			cells := s.optimizeRow(rowCells, row, grid.Cursor)
			tokens := MapTokens(cells)
			contentRows[row].Tokens = tokens
			contentRows[row].Index = lineNumber
			grid.OptimizedRows[row] = cells
			grid.DirtyRows[row] = false
		} else {
			contentRows[row].Tokens = MapTokens(grid.OptimizedRows[row])
			contentRows[row].Index = lineNumber
		}
	}
	return contentRows
}

// Extract line number from gutter and return the remaining cells
func (s *Screen) trimGutter(row []*Cell) ([]*Cell, int) {
	// Extract line number from gutter (first 6 characters)
	lineNumber := -1
	gutterWidth := 6
	if len(row) >= gutterWidth {
		gutterText := ""
		for i := 0; i < gutterWidth && i < len(row); i++ {
			gutterText += row[i].Char
		}

		// Use regex to extract the line number
		// This pattern looks for one or more digits in the gutter text
		re := regexp.MustCompile(`\d+`)
		matches := re.FindAllString(gutterText, -1)
		if len(matches) > 0 {
			// Use the first match if there are multiple numbers
			if parsedNum, err := strconv.Atoi(matches[0]); err == nil {
				lineNumber = parsedNum
			}
		}
	}

	// Return the row without the gutter
	if len(row) <= gutterWidth {
		return []*Cell{}, lineNumber
	}

	return row[gutterWidth:], lineNumber
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
	gutterWidth := 6 // The width of the gutter we trimmed

	for col := 0; col < len(rowCells); col++ {
		cell := rowCells[col]
		// Only show cursor if we're on the cursor's row and column (adjusted for gutter)
		isCursor := cursor.Row == currentRow && cursor.Col == col+gutterWidth

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
