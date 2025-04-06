package neovim

import (
	"errors"

	"github.com/neovim/go-client/nvim"
)

func getBufferFiletype(winId int) (*string, error) {
	windows, error := NvimInstance.Windows()
	if error != nil {
		return nil, error
	}

	var foundWindow *nvim.Window
	for _, window := range windows {
		currentWinId, _ := extractWindowId(window)
		if currentWinId == winId {
			foundWindow = &window
		}
	}
	if foundWindow == nil {
		return nil, errors.New("couldnt find window")
	}
	buffer, error := NvimInstance.WindowBuffer(*foundWindow)

	if error != nil {
		return nil, error
	}

	var filetype string
	error = NvimInstance.BufferOption(buffer, "filetype", &filetype)

	if error != nil {
		return nil, error
	}
	if filetype == "" {
		var treesitterContext bool
		NvimInstance.WindowVar(*foundWindow, "treesitter_context", treesitterContext)
		filetype = "treesitter_context"
	}
	return &filetype, nil
}

func isHex(window *Window) bool {
	ft := ""
	if window.Filetype != nil {
		ft = *window.Filetype
	}
	return !(ft == "fzflua_backdrop")
}

func (s *Screen) renderFloatingWindow(window *Window) []ContentRow {
	if isHex(window) {
		return window.Grid.toHex()
	} else {
		return s.renderFzfLua(window.Grid)
	}
}

func (s *Screen) renderFzfLua(grid *Grid) []ContentRow {
	content := s.optimizeFloatingGrid(grid)
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
