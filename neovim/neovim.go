package neovim

import (
	"context"
	"fmt"
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

func CalculateGridSize(windowWidth, windowHeight int) (rows, cols int) {
	cellWidth := 12
	cellHeight := 28
	statusLineHeight := 36
	availableHeight := windowHeight - statusLineHeight
	cols = windowWidth / cellWidth
	rows = availableHeight / cellHeight
	return rows, cols
}

func StartListening(ctx context.Context) {
	width, height := Runtime.WindowGetSize(ctx)
	rows, cols := CalculateGridSize(width, height)
	utils.Log(fmt.Sprintf("StartListening initializing screen with width=%d height=%d rows=%d cols=%d", width, height, rows, cols))

	screen = NewScreen(ctx, cols, rows)
	Runtime.EventsOn(ctx, "resize", func(optionalData ...interface{}) {
		width, height := Runtime.WindowGetSize(ctx)
		rows, cols := CalculateGridSize(width, height)
		screen.Resize(cols, rows)
	})
	var err error
	// nvimCtx, _ := context.WithCancel(ctx)
	NvimInstance, err = nvim.NewChildProcess(
		nvim.ChildProcessCommand("nvim"),
		nvim.ChildProcessArgs("--embed", "/home/stefan/Projects/nvim-gui/neovim/neovim.go"),
		nvim.ChildProcessContext(ctx),
	)
	if err != nil {
		log.Println(err)
		return
	}
	Runtime.EventsOn(ctx, "get-highlights", screen.sendInitialHighlights)
	Runtime.EventsOn(ctx, "substitute-jump", HandleSubstituteJump)
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

	if err := NvimInstance.Serve(); err != nil {
		log.Fatal(err)
	}
	log.Println("listening terminating")
}

func isVisualMode(mode string) bool {
	return mode == "v" || mode == "V" || mode == "\x16" // Normal, line, and block visual modes
}
