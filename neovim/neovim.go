package neovim

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"

	"nvim-gui/utils"

	"github.com/neovim/go-client/nvim"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	currentMode      string
	NvimInstance     *nvim.Nvim
	fgColorClasses   = make(map[string]string)
	bgColorClasses   = make(map[string]string)
	idClasses        = make(map[int][]string)
	effectiveHlIds   = make(map[string]int)
	effectiveHlIdsMu sync.Mutex
)

var NvimScreen *Screen

func CalculateGridSize(windowWidth, windowHeight int) (rows, cols int) {
	cellWidth := 12
	cellHeight := 28
	statusLineHeight := 36
	availableHeight := windowHeight - statusLineHeight
	cols = windowWidth / cellWidth
	rows = availableHeight / cellHeight
	utils.Log(fmt.Sprintf("CalculateGridSize rows=%d cols=%d windowHeight=%d windowWidth=%d", rows, cols, windowHeight, windowWidth))
	return rows, cols
}

func StartListening(ctx context.Context) {
	var err error
	width, height := Runtime.WindowGetSize(ctx)
	rows, cols := CalculateGridSize(width, height)
	utils.Log(fmt.Sprintf("StartListening initializing screen with width=%d height=%d rows=%d cols=%d", width, height, rows, cols))
	NvimScreen = NewScreen(ctx, cols, rows)

	// Set up resize handler
	Runtime.EventsOn(ctx, "resize", func(optionalData ...interface{}) {
		width, height := Runtime.WindowGetSize(ctx)
		rows, cols := CalculateGridSize(width, height)
		NvimScreen.Resize(cols, rows)
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
	NvimInstance.SetVar("nvim_gui", true)
	NvimInstance.SetVar("nvim_gui_channel", NvimInstance.ChannelID())
	NvimInstance.SetVar("keymaps", keymaps)

	if err != nil {
		log.Println(err)
		nvimCancel()
		Runtime.Quit(ctx)
		return
	}

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
			// "ext_messages":   true,
		}

		err = NvimInstance.AttachUI(cols, rows, opts)
		if err != nil {
			utils.Log(err.Error())
			nvimCancel()
			return
		}

		NvimInstance.RegisterHandler("redraw", func(updates ...[]interface{}) {
			NvimScreen.handleRedraw(updates)
		})

		SetupKeymaps()

		NvimInstance.RegisterHandler("BufEnter", func(_ *nvim.Nvim, data []string) {
			assert(len(data) == 1, "BufEnter: malformed data")
			filepath := filepath.Base(data[0])
			Runtime.EventsEmit(NvimScreen.ctx, "BufEnter", filepath)
		})

		NvimInstance.RegisterHandler("TrekClosed", func(_ *nvim.Nvim, windowArgs []uint64) {
			utils.Log("TrekClosed args=", windowArgs)
			assert(len(windowArgs) == 3, "incorrect length windowIds")
			windowIds := utils.MapArray(windowArgs, func(arg uint64) int {
				return utils.ReflectToInt(arg)
			})
			NvimScreen.closeTrek(windowIds)
		})

		if err := NvimInstance.Serve(); err != nil {
			utils.Log(fmt.Sprintf("Neovim process terminated: %v\n%s", err, debug.Stack()))
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
