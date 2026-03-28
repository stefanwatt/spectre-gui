package neovim

import (
	"nvim-gui/rendering"

	"github.com/charmbracelet/log"
)

func isHex(window *Window) bool {
	ft := ""
	if window.Buffer != nil {
		ft = (*window.Buffer).Filetype
	}
	return ft == "blink-cmp-menu"
}

func (s *Screen) renderFloatingWindow(window *Window) []rendering.ContentRow {
	log.Debug("renderFloatingWindow")
	// if isHex(window) {
	// 	return window.Grid.toHex()
	// } else {
	return s.renderFzfLua(window.Grid)
	// }
}

func (s *Screen) renderFzfLua(grid *Grid) []rendering.ContentRow {
	log.Debug("renderFzfLua")
	payload := rendering.BuildContentPayload(rendering.ContentInput{
		WindowID:   0,
		Filetype:   "fzflua",
		BufNr:      0,
		CursorLine: -1,
		Grid:       s.toRenderingGridData(grid),
		Meta:       nil,
	})
	content := payload.Content
	return trimPerimeter(content)
}

func trimPerimeter(contentRows []rendering.ContentRow) []rendering.ContentRow {
	if len(contentRows) < 3 {
		// Return an empty slice if there are fewer than 3 rows
		return []rendering.ContentRow{}
	}

	// Remove first and last row
	contentRows = contentRows[1 : len(contentRows)-1]

	for i := range contentRows {
		if len(contentRows[i].Tokens) < 3 {
			// If a row has fewer than 3 cells, make it empty
			contentRows[i].Tokens = []*rendering.Token{}
		} else {
			// Remove first and last cell of the row
			contentRows[i].Tokens = contentRows[i].Tokens[1 : len(contentRows[i].Tokens)-1]
		}
	}

	return contentRows
}
