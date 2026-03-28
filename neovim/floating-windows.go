package neovim

import "github.com/charmbracelet/log"

func isHex(window *Window) bool {
	ft := ""
	if window.Buffer != nil {
		ft = (*window.Buffer).Filetype
	}
	return ft == "blink-cmp-menu"
}

func (s *Screen) renderFloatingWindow(window *Window) []ContentRow {
	log.Debug("renderFloatingWindow")
	// if isHex(window) {
	// 	return window.Grid.toHex()
	// } else {
	return s.renderFzfLua(window.Grid)
	// }
}

func (s *Screen) renderFzfLua(grid *Grid) []ContentRow {
	log.Debug("renderFzfLua")
	content := s.optimizeGrid(grid, "fzflua", 0, -1)
	return trimPerimeter(content)
}

func trimPerimeter(contentRows []ContentRow) []ContentRow {
	if len(contentRows) < 3 {
		// Return an empty slice if there are fewer than 3 rows
		return []ContentRow{}
	}

	// Remove first and last row
	contentRows = contentRows[1 : len(contentRows)-1]

	for i := range contentRows {
		if len(contentRows[i].Tokens) < 3 {
			// If a row has fewer than 3 cells, make it empty
			contentRows[i].Tokens = []*Token{}
		} else {
			// Remove first and last cell of the row
			contentRows[i].Tokens = contentRows[i].Tokens[1 : len(contentRows[i].Tokens)-1]
		}
	}

	return contentRows
}
