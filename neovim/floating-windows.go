package neovim

import (
	"errors"
	"strconv"
	"strings"

	"github.com/neovim/go-client/nvim"
)

func getBufferFiletype(winId int) (*string, error) {
	windows, error := NvimInstance.Windows()
	if error != nil {
		return nil, error
	}

	var foundWindow *nvim.Window
	for _, window := range windows {
		currentWinId, _ := strconv.Atoi(strings.Split(window.String(), ":")[1])
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

func (s *Screen) renderFloatingWindow(window *Window) [][]*Cell {
	if isHex(window) {
		return window.Grid.toHexGrid().Cells
	} else {
		return s.renderFzfLua(window.Grid)
	}
}

func (s *Screen) renderFzfLua(grid *Grid) [][]*Cell {
	content := s.optimizeGrid(grid)
	return trimPerimeter(content)
}

func trimPerimeter(cells [][]*Cell) [][]*Cell {
	if len(cells) < 3 {
		// Return an empty slice if there are fewer than 3 rows
		return [][]*Cell{}
	}

	// Remove first and last row
	cells = cells[1 : len(cells)-1]

	for i := range cells {
		if len(cells[i]) < 3 {
			// If a row has fewer than 3 cells, make it empty
			cells[i] = []*Cell{}
		} else {
			// Remove first and last cell of the row
			cells[i] = cells[i][1 : len(cells[i])-1]
		}
	}

	return cells
}
