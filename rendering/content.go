package rendering

import (
	"nvim-gui/utils"
	"regexp"
	"strings"
)

var quoteRe = regexp.MustCompile(`^(\s*>)+`)

type Token struct {
	Text      string `json:"text" msgpack:"text"`
	Classes   string `json:"classes" msgpack:"classes"`
	Highlight int    `json:"highlight" msgpack:"hl_group"`
}

type ColRange struct {
	StartCol int
	EndCol   int
}

type CodeBlockOpts struct {
	Position string `json:"position"` // "first", "middle", "last"
}

type MarkdownMetaProvider interface {
	HeadingLevel(bufNr, bufferLine int) int
	TableMetaForLine(bufNr, bufferLine int) *TableMeta
	ImageMetaForLine(bufNr, bufferLine int) *ImageMeta
	TaskMetaForLine(bufNr, bufferLine int) (checked bool, isTask bool)
	CodeBlockMetaForLine(bufNr, bufferLine int) *CodeBlockMeta
	InlineCodeRanges(bufNr, bufferLine int) []ColRange
}

type GridData struct {
	Height        int
	TopLine       int
	DirtyRows     []bool
	Cells         [][]*Cell
	OptimizedRows [][]*Cell
	CachedTokens  [][]*Token
	MarkdownOpts  map[int]*MarkdownOpts
	Cursor        CursorPosition
}

type ContentRow struct {
	Index        int           `json:"index" msgpack:"row"` // basically the line number
	Tokens       []*Token      `json:"tokens" msgpack:"tokens"`
	MarkdownOpts *MarkdownOpts `json:"markdownOpts,omitempty"`
	Dirty        bool          `json:"dirty,omitempty"`
}

func (cr *ContentRow) ToString() string {
	var sb strings.Builder
	for _, token := range cr.Tokens {
		sb.WriteString(token.Text)
	}
	return sb.String()
}

// TODO: create mapper
func fromCell(cell *Cell) *Token {
	return &Token{
		Text:      cell.Char,
		Classes:   cell.ClassesToString(),
		Highlight: cell.Highlight,
	}
}

func MapTokens(cells []*Cell) []*Token {
	return utils.MapArray(cells, func(cell *Cell) *Token {
		return fromCell(cell)
	})
}

// cursorLine is the 0-indexed buffer line of the cursor, or -1 if unknown.
// When the cursor falls within a table, table rendering is suppressed (conceal/reveal).
func OptimizeGrid(grid *GridData, filetype string, bufNr, cursorLine int, meta MarkdownMetaProvider) []ContentRow {
	contentRows := make([]ContentRow, grid.Height)
	// Ensure CachedTokens slice exists
	if len(grid.CachedTokens) != grid.Height {
		grid.CachedTokens = make([][]*Token, grid.Height)
	}
	for row := 0; row < grid.Height; row++ {
		// TODO: handle folds — folded lines cause line numbers to jump,
		// topline+row won't be correct when folds are present.
		lineNumber := grid.TopLine + row + 1 // 1-indexed buffer line
		bufferLine := grid.TopLine + row     // 0-indexed buffer line
		rowCells := grid.Cells[row]

		if grid.DirtyRows[row] {
			cells := OptimizeRow(rowCells, row, CursorPosition{Row: grid.Cursor.Row, Col: grid.Cursor.Col})
			tokens := MapTokens(cells)
			contentRows[row].Tokens = tokens
			contentRows[row].Index = lineNumber
			contentRows[row].Dirty = true
			grid.OptimizedRows[row] = cells
			grid.CachedTokens[row] = tokens
			grid.DirtyRows[row] = false

			if filetype == "markdown" && meta != nil {
				markdownOpts := getMarkdownOpts(contentRows[row])
				if headingLevel := meta.HeadingLevel(bufNr, bufferLine); headingLevel > 0 {
					markdownOpts.HeadingLevel = headingLevel
				}
				tableMeta := meta.TableMetaForLine(bufNr, bufferLine)
				if tableMeta != nil {
					markdownOpts.Table = buildTableRowOpts(tableMeta, bufferLine, contentRows[row].Tokens)
				}
				imageMeta := meta.ImageMetaForLine(bufNr, bufferLine)
				if imageMeta != nil {
					markdownOpts.Image = &ImageOpts{URL: imageMeta.URL, AltText: imageMeta.AltText}
				}
				if checked, isTask := meta.TaskMetaForLine(bufNr, bufferLine); isTask && cursorLine != bufferLine {
					markdownOpts.Task = &TaskOpts{Checked: checked}
				}
				if cbMeta := meta.CodeBlockMetaForLine(bufNr, bufferLine); cbMeta != nil {
					position := "middle"
					if bufferLine == cbMeta.StartLine {
						position = "first"
					} else if bufferLine == cbMeta.EndLine-1 {
						position = "last"
					}
					markdownOpts.CodeBlock = &CodeBlockOpts{Position: position}
				}
				if inlineRanges := meta.InlineCodeRanges(bufNr, bufferLine); len(inlineRanges) > 0 && cursorLine != bufferLine {
					applyInlineCodeClass(contentRows[row].Tokens, inlineRanges)
				}
				contentRows[row].MarkdownOpts = markdownOpts
				grid.MarkdownOpts[row] = markdownOpts
			}
		} else {
			// Reuse cached tokens for non-dirty rows (avoids allocating new Token objects)
			if grid.CachedTokens[row] != nil {
				contentRows[row].Tokens = grid.CachedTokens[row]
			} else {
				contentRows[row].Tokens = MapTokens(grid.OptimizedRows[row])
			}
			contentRows[row].Index = lineNumber
			if filetype == "markdown" && meta != nil {
				// Must recompute table opts even for non-dirty rows because TopLine changes on scroll
				markdownOpts := getMarkdownOpts(contentRows[row])
				if headingLevel := meta.HeadingLevel(bufNr, bufferLine); headingLevel > 0 {
					markdownOpts.HeadingLevel = headingLevel
				}
				tableMeta := meta.TableMetaForLine(bufNr, bufferLine)
				if tableMeta != nil {
					markdownOpts.Table = buildTableRowOpts(tableMeta, bufferLine, contentRows[row].Tokens)
				}
				imageMeta := meta.ImageMetaForLine(bufNr, bufferLine)
				if imageMeta != nil {
					markdownOpts.Image = &ImageOpts{URL: imageMeta.URL, AltText: imageMeta.AltText}
				}
				if checked, isTask := meta.TaskMetaForLine(bufNr, bufferLine); isTask && cursorLine != bufferLine {
					markdownOpts.Task = &TaskOpts{Checked: checked}
				}
				if cbMeta := meta.CodeBlockMetaForLine(bufNr, bufferLine); cbMeta != nil {
					position := "middle"
					if bufferLine == cbMeta.StartLine {
						position = "first"
					} else if bufferLine == cbMeta.EndLine-1 {
						position = "last"
					}
					markdownOpts.CodeBlock = &CodeBlockOpts{Position: position}
				}
				if inlineRanges := meta.InlineCodeRanges(bufNr, bufferLine); len(inlineRanges) > 0 && cursorLine != bufferLine {
					applyInlineCodeClass(contentRows[row].Tokens, inlineRanges)
				}
				contentRows[row].MarkdownOpts = markdownOpts
				grid.MarkdownOpts[row] = markdownOpts
			} else if opts, exists := grid.MarkdownOpts[row]; exists {
				contentRows[row].MarkdownOpts = opts
			}
		}
	}
	return contentRows
}

func applyInlineCodeClass(tokens []*Token, ranges []ColRange) {
	col := 0
	for _, token := range tokens {
		tokenLen := len([]rune(token.Text))
		tokenEnd := col + tokenLen
		for _, r := range ranges {
			if col < r.EndCol && tokenEnd > r.StartCol {
				token.Classes += " code"
				// Strip backtick delimiters at range boundaries
				if col <= r.StartCol {
					token.Text = strings.TrimLeft(token.Text, "`")
				}
				if tokenEnd >= r.EndCol {
					token.Text = strings.TrimRight(token.Text, "`")
				}
				break
			}
		}
		col = tokenEnd
	}
}

func getMarkdownOpts(contentRow ContentRow) *MarkdownOpts {
	text := contentRow.ToString()
	count := 0
	match := quoteRe.FindString(text)
	for _, r := range match {
		if r == '>' {
			count++
		}
	}
	return &MarkdownOpts{QuoteLevel: count}
}

// isPipeSeparator checks if a character is a pipe or box-drawing vertical line
func isPipeSeparator(r rune) bool {
	return r == '|' || r == '│' || r == '┃'
}

// splitAtPipes splits a string at pipe characters (ASCII or box-drawing)
func splitAtPipes(text string) []string {
	var parts []string
	var current strings.Builder
	for _, r := range text {
		if isPipeSeparator(r) {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}
	parts = append(parts, current.String())
	return parts
}

// splitTokensIntoCells splits a token stream at pipe characters into separate cell groups.
// It filters out leading/trailing empty cells (from outer pipes) and trims whitespace from cell tokens.
// Handles both ASCII pipes (|) and box-drawing vertical lines (│).
func splitTokensIntoCells(tokens []*Token) [][]*Token {
	var cells [][]*Token
	var current []*Token

	for _, token := range tokens {
		text := token.Text
		parts := splitAtPipes(text)
		for i, part := range parts {
			if i > 0 {
				// Pipe boundary — flush current cell
				cells = append(cells, current)
				current = nil
			}
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				current = append(current, &Token{
					Text:      trimmed,
					Classes:   token.Classes,
					Highlight: token.Highlight,
				})
			}
		}
	}
	// Flush last cell
	if len(current) > 0 {
		cells = append(cells, current)
	}

	// Filter out empty cells (from leading/trailing pipes)
	var filtered [][]*Token
	for _, cell := range cells {
		if len(cell) > 0 {
			filtered = append(filtered, cell)
		}
	}
	return filtered
}

// isSeparatorRow checks if the row text looks like a table separator
// Handles both ASCII (|---|---|) and box-drawing (├───┼───┤) formats.
func isSeparatorRow(tokens []*Token) bool {
	var sb strings.Builder
	for _, t := range tokens {
		sb.WriteString(t.Text)
	}
	text := strings.TrimSpace(sb.String())
	if text == "" {
		return false
	}
	// Check if it contains dashes or box-drawing horizontal lines
	hasDash := strings.ContainsAny(text, "-─━")
	if !hasDash {
		return false
	}
	// Remove all expected separator characters
	for _, r := range text {
		switch {
		case r == '-' || r == ':' || r == ' ':
			continue
		case isPipeSeparator(r):
			continue
		// Box-drawing characters used in separator rows
		case r == '─' || r == '━' || r == '├' || r == '┤' || r == '┼' || r == '╋':
			continue
		default:
			return false
		}
	}
	return true
}

// isBorderRow checks if the row is a top/bottom box-drawing border (┌─┬─┐ or └─┴─┘)
func isBorderRow(tokens []*Token) bool {
	var sb strings.Builder
	for _, t := range tokens {
		sb.WriteString(t.Text)
	}
	text := strings.TrimSpace(sb.String())
	if text == "" {
		return false
	}
	for _, r := range text {
		switch r {
		case '┌', '┐', '└', '┘', '┬', '┴', '─', '━', ' ':
			continue
		default:
			return false
		}
	}
	return true
}

func OptimizeRow(rowCells []*Cell, currentRow int, cursor CursorPosition) []*Cell {
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
			for class := range cell.Classes {
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
