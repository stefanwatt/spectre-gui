package neovim

import (
	"strings"

	"github.com/akiyosi/goneovim/util"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ContentRow struct {
	Index  int     `json:"index"`
	Tokens []*Cell `json:"tokens"`
}

func (s *Screen) optimizeGrid(grid *Grid) []ContentRow {
	contentRows := make([]ContentRow, grid.Height)
	for row := 0; row < grid.Height; row++ {
		if grid.DirtyRows[row] {
			contentRows[row].Tokens = s.optimizeRow(grid, row)
			contentRows[row].Index = row
			grid.OptimizedRows[row] = contentRows[row].Tokens
			grid.DirtyRows[row] = false
		} else {
			contentRows[row].Tokens = grid.OptimizedRows[row]
		}
	}
	return contentRows
}

func (s *Screen) optimizeRow(grid *Grid, row int) []*Cell {
	optimizedRow := make([]*Cell, 0)
	if row >= len(grid.Cells) {
		return optimizedRow
	}

	currentRow := grid.Cells[row]
	cursor := grid.Cursor
	var currentToken *Cell
	lastHl := 0

	for col, cell := range currentRow {
		isCursor := cursor.Row == row && cursor.Col == col

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

func (s *Screen) handleCmdlineShow(args []interface{}) {
	arg := args[0].([]interface{})

	content := ""
	contentChunks := arg[0].([]interface{})
	for _, e := range contentChunks {
		a := e.([]interface{})

		if len(a) < 2 {
			// content += a[0].(string)
			content += strings.Replace(a[0].(string), "\t", " ", -1)
		} else {
			if len(contentChunks) == 1 {
				// content += a[1].(string)
				content += strings.Replace(a[1].(string), "\t", " ", -1)
			} else {
				content +=
					sanitize(a[1].(string))
			}
		}
	}
	// content := arg[0].([]interface{})[0].([]interface{})[1].(string)

	pos := util.ReflectToInt(arg[1])
	firstc := arg[2].(string)
	prompt := arg[3].(string)
	indent := util.ReflectToInt(arg[4])
	// level := util.ReflectToInt(arg[5])
	// fmt.Println("cmdline show", content, pos, firstc, prompt, indent, level)

	Runtime.EventsEmit(s.ctx, "cmdline_show", map[string]interface{}{
		"content": content,
		"pos":     pos,
		"firstc":  firstc,
		"prompt":  prompt,
		"indent":  indent,
	})
}
