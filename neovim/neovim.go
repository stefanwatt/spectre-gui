package neovim

import (
	"context"
	"fmt"
	"log"
	"os"

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
	colorClasses map[string]string
	idClasses    map[int][]string
	effectiveHlIds map[string]int = make(map[string]int)
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
	var err error
	go loadDatabase()
	go processHlAttrQueue()
	waitForHlAttrDefine()
	width, height := Runtime.WindowGetSize(ctx)
	rows, cols := CalculateGridSize(width, height)
	utils.Log(fmt.Sprintf("StartListening initializing screen with width=%d height=%d rows=%d cols=%d", width, height, rows, cols))
	screen = NewScreen(ctx, cols, rows)

	// Set up resize handler
	Runtime.EventsOn(ctx, "resize", func(optionalData ...interface{}) {
		width, height := Runtime.WindowGetSize(ctx)
		rows, cols := CalculateGridSize(width, height)
		screen.Resize(cols, rows)
	})

	nvimCtx, nvimCancel := context.WithCancel(ctx)
	nvimExitChan := make(chan struct{})

	var nvimArgs nvim.ChildProcessOption
	if len(os.Args) > 1 {
		filepath := os.Args[1]
		nvimArgs = nvim.ChildProcessArgs("--embed", filepath)
	} else {
		nvimArgs = nvim.ChildProcessArgs("--embed")
	}

	NvimInstance, err = nvim.NewChildProcess(
		nvim.ChildProcessCommand("nvim"),
		nvimArgs,
		nvim.ChildProcessContext(nvimCtx),
	)

	if err != nil {
		log.Println(err)
		nvimCancel()
		Runtime.Quit(ctx)
		return
	}
	// err = NvimInstance.SetOption("relativenumber", false)

	Runtime.EventsOn(ctx, "get-highlights", updateHighlightCSS)
	Runtime.EventsOn(ctx, "substitute-jump", HandleSubstituteJump)

	// Run a goroutine to handle Neovim serving and exit
	go func() {
		defer close(nvimExitChan)
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
			nvimCancel()
			return
		}

		NvimInstance.RegisterHandler("redraw", func(updates ...[]interface{}) {
			screen.handleRedraw(updates)
		})

		if err := NvimInstance.Serve(); err != nil {
			utils.Log(fmt.Sprintf("Neovim process terminated: %v", err))
		}

		// Neovim has exited, signal to quit the app
		utils.Log("Neovim process has terminated, quitting application")
	}()

	// Wait for Neovim to exit or context to be cancelled
	select {
	case <-nvimExitChan:
		// Neovim exited, quit the app
		utils.Log("Detected Neovim exit, shutting down application")
		Runtime.Quit(ctx)
	case <-ctx.Done():
		// Parent context was cancelled
		nvimCancel()
	}

	log.Println("listening terminating")
}

func isVisualMode(mode string) bool {
	return mode == "v" || mode == "V" || mode == "\x16" // Normal, line, and block visual modes
}
