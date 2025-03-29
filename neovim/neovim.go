package neovim

import (
	"context"
	"log"

	"nvim-gui/utils"

	"github.com/neovim/go-client/nvim"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type CursorState struct {
	Row uint64
	Col uint64
}

var (
	cursorState  CursorState
	currentMode  string
	NvimInstance *nvim.Nvim
)

var screen *Screen

func StartListening(ctx context.Context) {
	cols := 140
	rows := 57
	screen = NewScreen(ctx, cols, rows)
	var err error
	NvimInstance, err = nvim.NewChildProcess(
		nvim.ChildProcessCommand("nvim"),
		nvim.ChildProcessArgs("--embed", "/home/stefan/Projects/nvim-gui/neovim/screen.go"),
		nvim.ChildProcessContext(context.Background()),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer NvimInstance.Close()

	opts := map[string]interface{}{
		"rgb":            true,
		"ext_linegrid":   true,
		"ext_multigrid":  true,
		"ext_hlstate":    true,
		"ext_termcolors": true,
		"ext_cmdline":    true,
		"ext_popupmenu":  true,
		"ext_tabline":    true,
		"ext_messages":   true,
	}
	err = NvimInstance.AttachUI(cols, rows, opts)
	if err != nil {
		utils.Log(err.Error())
	}
		
	NvimInstance.RegisterHandler("redraw", func(updates ...[]interface{}) {
		screen.handleRedraw(updates)
	})

	NvimInstance.RegisterHandler("nvim-gui-cursor-moved", func(v *nvim.Nvim, cursor_move_event CursorMoveEvent) {
		cursorState.Row = cursor_move_event.Row
		cursorState.Col = cursor_move_event.Col
		UpdateCursor(ctx, cursor_move_event)
	})

	NvimInstance.RegisterHandler("nvim-gui-mode-changed", func(v *nvim.Nvim, args []string) {
		mode := args[0]
		currentMode = mode
		Runtime.EventsEmit(ctx, "mode-changed", mode)
		
	})

	if err := NvimInstance.Serve(); err != nil {
		log.Fatal(err)
	}
	log.Println("listening terminating")
}

func isVisualMode(mode string) bool {
	return mode == "v" || mode == "V" || mode == "\x16" // Normal, line, and block visual modes
}
