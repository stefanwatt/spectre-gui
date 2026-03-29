package neovim

import (
	"errors"
	"fmt"
	"nvim-gui/utils"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/neovim/go-client/nvim"
)

// Window represents a Neovim window, which can be a regular window or a floating window
type Window struct {
	ID                  int // Window ID
	Dirty               bool
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
	IsFileExplorer      bool   // Whether this is a mini.files directory window
	Title               string // Floating window title (e.g. directory path for mini.files)
	Hidden              bool
	lineNumbers         bool
	relativeLineNumbers bool
	Buffer              *Buffer
	Cursor              *Cursor
	Mode                string
}

// NewWindow creates a new window with the given ID
func NewWindow(id int) *Window {
	return &Window{
		ID:                  id,
		Dirty:               true,
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
	log.Debug(fmt.Sprintf("GetWindow winId=%d", winId))
	nvimWindows, err := NvimClient.Windows()
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

func (window *Window) Resize(width, height int) {
	window.Width = width
	window.Height = height
	window.Dirty = true

	// If this is a small 1x1 window, it's probably not a completion window
	if width == 1 && height == 1 {
		window.IsPopupmenu = false
	}
}

func resizeFloatingWindow(winId, cols, rows int) error {
	win, err := getWindow(winId)
	if err != nil {
		return err
	}
	config, err := NvimClient.WindowConfig(*win)
	if err != nil {
		return err
	}
	config.Width = cols
	config.Height = rows
	return NvimClient.SetWindowConfig(*win, config)
}
