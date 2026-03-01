package neovim

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
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

	// Handle requests from frontend to re-emit current state (for late-connecting clients like Playwright)
	Runtime.EventsOn(ctx, "request-state", func(optionalData ...interface{}) {
		if NvimScreen != nil {
			NvimScreen.EmitCurrentState()
		}
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

		// Register all handlers BEFORE AttachUI so no events are missed.
		// The go-client's internal goroutine dispatches notifications immediately;
		// if handlers aren't registered when hl_attr_define arrives, highlights are lost.
		NvimInstance.RegisterHandler("redraw", func(updates ...[]interface{}) {
			NvimScreen.handleRedraw(updates)
		})

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

		NvimInstance.RegisterHandler("MarkdownTables", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 2 {
					continue
				}
				bufNr := utils.ReflectToInt(data[0])
				tablesRaw, ok := data[1].([]interface{})
				if !ok {
					continue
				}
				tables := parseMarkdownTables(tablesRaw)
				utils.Log(fmt.Sprintf("MarkdownTables received: bufNr=%d, tables=%d", bufNr, len(tables)))
				NvimScreen.setTableMetadata(bufNr, tables)
			}
		})

		NvimInstance.RegisterHandler("MarkdownImages", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 2 {
					continue
				}
				bufNr := utils.ReflectToInt(data[0])
				imagesRaw, ok := data[1].([]interface{})
				if !ok {
					continue
				}
				images := parseMarkdownImages(imagesRaw)
				utils.Log(fmt.Sprintf("MarkdownImages received: bufNr=%d, images=%d", bufNr, len(images)))
				NvimScreen.setImageMetadata(bufNr, images)
			}
		})

		NvimInstance.RegisterHandler("MarkdownHeadings", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 2 {
					continue
				}
				bufNr := utils.ReflectToInt(data[0])
				headingsRaw, ok := data[1].([]interface{})
				if !ok {
					continue
				}
				headings := parseMarkdownHeadings(headingsRaw)
				NvimScreen.setHeadingMetadata(bufNr, headings)
			}
		})

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

		// Read colorcolumn value and highlight color before disabling
		readColorColumn(NvimScreen)

		// Disable neovim's gutter, wrapping, and colorcolumn — we render these ourselves.
		NvimInstance.Command("set nonumber norelativenumber signcolumn=no foldcolumn=0 nowrap colorcolumn=")

		// Open test file if env var is set (used by e2e tests)
		if testFile := os.Getenv("NVIM_GUI_TEST_FILE"); testFile != "" {
			// Convert to absolute path if relative
			absPath := testFile
			if !filepath.IsAbs(testFile) {
				cwd, err := os.Getwd()
				if err == nil {
					absPath = filepath.Join(cwd, testFile)
				}
			}
			utils.Log(fmt.Sprintf("Opening test file: %s", absPath))
			err := NvimInstance.Command(fmt.Sprintf("edit %s", absPath))
			if err != nil {
				utils.Log(fmt.Sprintf("Error opening test file: %v", err))
			}
		}

		SetupKeymaps()

		// Set up markdown table detection via treesitter
		if err := NvimInstance.ExecLua(markdownTablesLua, nil, NvimInstance.ChannelID()); err != nil {
			utils.Log(fmt.Sprintf("Error loading markdown tables Lua: %v", err))
		}

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

func readColorColumn(screen *Screen) {
	// Read colorcolumn setting (returns a list of strings like {"80"} or {"+1", "120"})
	var ccValues []interface{}
	err := NvimInstance.ExecLua("return vim.opt.colorcolumn:get()", &ccValues)
	if err != nil {
		utils.Log(fmt.Sprintf("Error reading colorcolumn: %v", err))
		return
	}

	var columns []int
	for _, v := range ccValues {
		str := fmt.Sprintf("%v", v)
		// Only support absolute column numbers (skip relative like "+1")
		col, err := strconv.Atoi(str)
		if err == nil && col > 0 {
			columns = append(columns, col)
		}
	}
	screen.ColorColumns = columns

	if len(columns) == 0 {
		return
	}

	// Read ColorColumn highlight group color
	var hlResult map[string]interface{}
	err = NvimInstance.ExecLua("return vim.api.nvim_get_hl(0, {name='ColorColumn'})", &hlResult)
	if err != nil {
		utils.Log(fmt.Sprintf("Error reading ColorColumn highlight: %v", err))
		return
	}

	if bg, ok := hlResult["bg"]; ok {
		bgInt := utils.ReflectToInt(bg)
		screen.ColorColumnColor = fmt.Sprintf("#%06x", bgInt)
	}
	utils.Log(fmt.Sprintf("ColorColumn: columns=%v color=%s", screen.ColorColumns, screen.ColorColumnColor))
}

func isVisualMode(mode string) bool {
	return mode == "v" || mode == "V" || mode == "\x16" // Normal, line, and block visual modes
}
