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

type TableMeta struct {
	StartLine  int      // 0-indexed buffer line
	EndLine    int      // exclusive
	Alignments []string // "left", "center", "right"
}

type TableRowOpts struct {
	TableID    int        `json:"tableId"`
	RowType    string     `json:"rowType"` // "header", "separator", "data"
	Cells      [][]*Token `json:"cells"`
	Alignments []string   `json:"alignments"`
}

type ImageMeta struct {
	Line    int    // 0-indexed buffer line
	URL     string
	AltText string
}

type ImageOpts struct {
	URL     string `json:"url"`
	AltText string `json:"altText"`
}

type TaskOpts struct {
	Checked bool `json:"checked"`
}

type CodeBlockMeta struct {
	StartLine int // 0-indexed buffer line (inclusive)
	EndLine   int // 0-indexed buffer line (exclusive)
}

type ColRange struct {
	StartCol int
	EndCol   int
}

type CodeBlockOpts struct {
	Position string `json:"position"` // "first", "middle", "last"
}


type MarkdownOpts struct {
	QuoteLevel   int            `json:"quoteLevel"`
	HeadingLevel int            `json:"headingLevel,omitempty"`
	Table        *TableRowOpts  `json:"table,omitempty"`
	Image        *ImageOpts     `json:"image,omitempty"`
	Task         *TaskOpts      `json:"task,omitempty"`
	CodeBlock    *CodeBlockOpts `json:"codeBlock,omitempty"`
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

// cursorLine is the 0-indexed buffer line of the cursor, or -1 if unknown.
// When the cursor falls within a table, table rendering is suppressed (conceal/reveal).
func (s *Screen) optimizeGrid(grid *Grid, filetype string, bufNr int, cursorLine int) []ContentRow {
	contentRows := make([]ContentRow, grid.Height)
	for row := 0; row < grid.Height; row++ {
		// TODO: handle folds — folded lines cause line numbers to jump,
		// topline+row won't be correct when folds are present.
		lineNumber := grid.TopLine + row + 1 // 1-indexed buffer line
		bufferLine := grid.TopLine + row      // 0-indexed buffer line
		rowCells := grid.Cells[row]

		if grid.DirtyRows[row] {
			cells := s.optimizeRow(rowCells, row, grid.Cursor)
			tokens := MapTokens(cells)
			contentRows[row].Tokens = tokens
			contentRows[row].Index = lineNumber
			grid.OptimizedRows[row] = cells
			grid.DirtyRows[row] = false

			if filetype == "markdown" {
				markdownOpts := getMarkdownOpts(contentRows[row], cursorLine, bufferLine)
				if headingLevel := s.getHeadingLevel(bufNr, bufferLine); headingLevel > 0 && cursorLine != bufferLine {
					markdownOpts.HeadingLevel = headingLevel
				}
				tableMeta := s.getTableMetaForLine(bufNr, bufferLine)
				if tableMeta != nil && !isCursorInTable(cursorLine, tableMeta) {
					markdownOpts.Table = buildTableRowOpts(tableMeta, bufferLine, contentRows[row].Tokens)
				}
				imageMeta := s.getImageMetaForLine(bufNr, bufferLine)
				if imageMeta != nil && cursorLine != bufferLine {
					markdownOpts.Image = &ImageOpts{URL: imageMeta.URL, AltText: imageMeta.AltText}
				}
				if checked, isTask := s.getTaskMetaForLine(bufNr, bufferLine); isTask && cursorLine != bufferLine {
					markdownOpts.Task = &TaskOpts{Checked: checked}
				}
				if cbMeta := s.getCodeBlockMetaForLine(bufNr, bufferLine); cbMeta != nil && !isCursorInCodeBlock(cursorLine, cbMeta) {
					position := "middle"
					if bufferLine == cbMeta.StartLine {
						position = "first"
					} else if bufferLine == cbMeta.EndLine-1 {
						position = "last"
					}
					markdownOpts.CodeBlock = &CodeBlockOpts{Position: position}
				}
				if inlineRanges := s.getInlineCodeRanges(bufNr, bufferLine); len(inlineRanges) > 0 && cursorLine != bufferLine {
					applyInlineCodeClass(contentRows[row].Tokens, inlineRanges)
				}
				contentRows[row].MarkdownOpts = markdownOpts
				grid.MarkdownOpts[row] = markdownOpts
			}
		} else {
			contentRows[row].Tokens = MapTokens(grid.OptimizedRows[row])
			contentRows[row].Index = lineNumber
			if filetype == "markdown" {
				// Must recompute table opts even for non-dirty rows because TopLine changes on scroll
				markdownOpts := getMarkdownOpts(contentRows[row], cursorLine, bufferLine)
				if headingLevel := s.getHeadingLevel(bufNr, bufferLine); headingLevel > 0 && cursorLine != bufferLine {
					markdownOpts.HeadingLevel = headingLevel
				}
				tableMeta := s.getTableMetaForLine(bufNr, bufferLine)
				if tableMeta != nil && !isCursorInTable(cursorLine, tableMeta) {
					markdownOpts.Table = buildTableRowOpts(tableMeta, bufferLine, contentRows[row].Tokens)
				}
				imageMeta := s.getImageMetaForLine(bufNr, bufferLine)
				if imageMeta != nil && cursorLine != bufferLine {
					markdownOpts.Image = &ImageOpts{URL: imageMeta.URL, AltText: imageMeta.AltText}
				}
				if checked, isTask := s.getTaskMetaForLine(bufNr, bufferLine); isTask && cursorLine != bufferLine {
					markdownOpts.Task = &TaskOpts{Checked: checked}
				}
				if cbMeta := s.getCodeBlockMetaForLine(bufNr, bufferLine); cbMeta != nil && !isCursorInCodeBlock(cursorLine, cbMeta) {
					position := "middle"
					if bufferLine == cbMeta.StartLine {
						position = "first"
					} else if bufferLine == cbMeta.EndLine-1 {
						position = "last"
					}
					markdownOpts.CodeBlock = &CodeBlockOpts{Position: position}
				}
				if inlineRanges := s.getInlineCodeRanges(bufNr, bufferLine); len(inlineRanges) > 0 && cursorLine != bufferLine {
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

// isCursorInTable returns true if the cursor's 0-indexed buffer line
// falls within the table's range (with the same ±1 border extension
// used by getTableMetaForLine for box-drawing rows).
func isCursorInTable(cursorLine int, meta *TableMeta) bool {
	if cursorLine < 0 || meta == nil {
		return false
	}
	return cursorLine >= meta.StartLine-1 && cursorLine < meta.EndLine+1
}

func isCursorInCodeBlock(cursorLine int, meta *CodeBlockMeta) bool {
	if cursorLine < 0 || meta == nil {
		return false
	}
	return cursorLine >= meta.StartLine && cursorLine < meta.EndLine
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

func getMarkdownOpts(contentRow ContentRow, cursorLine int, bufferLine int) *MarkdownOpts {
	text := contentRow.ToString()
	count := 0
	// Only set quoteLevel when cursor is not on this line
	if cursorLine != bufferLine {
		re := regexp.MustCompile(`^(\s*>)+`)
		match := re.FindString(text)
		for _, r := range match {
			if r == '>' {
				count++
			}
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

func buildTableRowOpts(meta *TableMeta, bufferLine int, tokens []*Token) *TableRowOpts {
	// Detect row type based on content, not just position
	// Box-drawing renderers may add border rows and change separators
	if isBorderRow(tokens) {
		return &TableRowOpts{
			TableID:    meta.StartLine,
			RowType:    "separator",
			Cells:      nil,
			Alignments: meta.Alignments,
		}
	}
	if isSeparatorRow(tokens) {
		return &TableRowOpts{
			TableID:    meta.StartLine,
			RowType:    "separator",
			Cells:      nil,
			Alignments: meta.Alignments,
		}
	}

	cells := splitTokensIntoCells(tokens)
	rowType := "data"
	if bufferLine == meta.StartLine {
		rowType = "header"
	}
	return &TableRowOpts{
		TableID:    meta.StartLine,
		RowType:    rowType,
		Cells:      cells,
		Alignments: meta.Alignments,
	}
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
