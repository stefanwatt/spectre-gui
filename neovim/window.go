package neovim

import (
	"errors"
	"fmt"
	"nvim-gui/utils"
	"strconv"
	"strings"

	"github.com/neovim/go-client/nvim"
)

// Window represents a Neovim window, which can be a regular window or a floating window
type Window struct {
	ID                  int // Window ID
	Dirty               bool
	Grid                *Grid
	Type                string // "normal" or "floating"
	Anchor              string // Anchor position for floating windows
	AnchorGrid          int    // Grid ID this window is anchored to
	StartRow            int    // Row position
	StartCol            int    // Column position
	Width               int    // Window width
	Height              int    // Window height
	Focusable           bool   // Whether the window can be focused
	ZIndex              int    // Z-index for floating windows
	IsPopupmenu         bool   // Whether this is a completion menu
	Hidden              bool
	lineNumbers         bool
	relativeLineNumbers bool
	Buffer              *Buffer
	Cursor              *Cursor
	Mode                string
}

type WindowAPI struct {
	ID                  int     `json:"id"`
	Type                string  `json:"type"`
	Width               float64 `json:"width"`  // in percent of screen
	Height              float64 `json:"height"` // in percent of screen
	ColStart            int     `json:"colStart"`
	ColEnd              int     `json:"colEnd"`
	RowStart            int     `json:"rowStart"`
	RowEnd              int     `json:"rowEnd"`
	LineNumbers         bool    `json:"lineNumbers"`
	RelativeLineNumbers bool    `json:"relativeLineNumbers"`
	Filetype            string  `json:"filetype"`
	Filepath            string  `json:"filepath"`
	Mode                string  `json:"mode"`
	Cursor              *Cursor `json:"cursor"`
}

// Equal compares two WindowAPI structs and returns true if they are equal.
func (w *WindowAPI) Equal(other *WindowAPI) bool {
	if other == nil {
		return false
	}

	return w.ID == other.ID &&
		w.Type == other.Type &&
		w.Width == other.Width &&
		w.Height == other.Height &&
		w.ColStart == other.ColStart &&
		w.ColEnd == other.ColEnd &&
		w.RowStart == other.RowStart &&
		w.RowEnd == other.RowEnd
}

// NewWindow creates a new window with the given ID and grid ID
func NewWindow(id int, grid *Grid) *Window {
	return &Window{
		ID:                  id,
		Dirty:               true,
		Grid:                grid,
		Type:                "normal",
		Anchor:              "",
		AnchorGrid:          0,
		StartRow:            0,
		StartCol:            0,
		Width:               0,
		Height:              0,
		Focusable:           true,
		ZIndex:              0,
		IsPopupmenu:         false,
		Cursor:              &Cursor{Row: 1, Col: 1},
		Mode:                "normal",
		lineNumbers:         true,
		relativeLineNumbers: true,
	}
}

func (w *Window) IsFloating() bool {
	return w.Type == "floating"
}

func extractWindowId(window nvim.Window) (int, error) {
	return strconv.Atoi(strings.Split(window.String(), ":")[1])
}

func getWindow(winId int) (*nvim.Window, error) {
	if winId < 1 {
		return nil, errors.New(fmt.Sprintf("invalid winId %d", winId))
	}
	utils.Log(fmt.Sprintf("GetWindow winId=%d", winId))
	nvimWindows, err := NvimInstance.Windows()
	if err != nil {
		return nil, err
	}
	foundWin, err := utils.Find(nvimWindows, func(win nvim.Window) bool {
		currentWinId, extractErr := extractWindowId(win)
		return extractErr == nil && currentWinId == winId
	})
	if err != nil {
		return nil, err
	}
	return &foundWin, err
}
