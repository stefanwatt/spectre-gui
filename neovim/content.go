package neovim

import (
	"regexp"
	"strconv"
	"strings"
)

type ContentRow struct {
	Index  int     `json:"index"`
	Tokens []*Cell `json:"tokens"`
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
			tokens := s.optimizeRow(rowCells, row, grid.Cursor)
			contentRows[row].Tokens = tokens
			contentRows[row].Index = lineNumber
			grid.OptimizedRows[row] = tokens
			grid.DirtyRows[row] = false
		} else {
			contentRows[row].Tokens = grid.OptimizedRows[row]
			contentRows[row].Index = lineNumber
		}
	}
	return contentRows
}

func (s *Screen) optimizeFloatingGrid(grid *Grid) []ContentRow {
	contentRows := make([]ContentRow, grid.Height)
	for row := 0; row < grid.Height; row++ {
		rowCells := grid.Cells[row]
		if grid.DirtyRows[row] {
			// Pass the current row number to optimizeRow
			tokens := s.optimizeRow(rowCells, row, grid.Cursor)
			contentRows[row].Tokens = tokens
			contentRows[row].Index = row
			grid.OptimizedRows[row] = tokens
			grid.DirtyRows[row] = false
		} else {
			contentRows[row].Tokens = grid.OptimizedRows[row]
			contentRows[row].Index = row
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

// Now optimizeRow works with just a single row and cursor position
// Added currentRow parameter to correctly check cursor position
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
			cursorCell := &Cell{
				Char:      cell.Char,
				Highlight: cell.Highlight,
				Classes:   "cursor",
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
