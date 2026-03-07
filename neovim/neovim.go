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
	// Use default grid size for initial creation
	// The frontend will call OnResize() once mounted
	rows, cols := 24, 80
	utils.Log(fmt.Sprintf("StartListening initializing screen with default rows=%d cols=%d", rows, cols))
	NvimScreen = NewScreen(ctx, cols, rows, App)

	// Note: resize and request-state handlers are now service methods on App:
	// - App.OnResize(width, height) called from frontend
	// - App.RequestState() called from frontend

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
		if App != nil {
			App.Quit()
		}
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
			if App != nil {
			App.Event.Emit("BufEnter", filepath)
		}
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

		NvimInstance.RegisterHandler("MarkdownTasks", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 2 {
					continue
				}
				bufNr := utils.ReflectToInt(data[0])
				tasksRaw, ok := data[1].([]interface{})
				if !ok {
					continue
				}
				tasks := parseMarkdownTasks(tasksRaw)
				NvimScreen.setTaskMetadata(bufNr, tasks)
			}
		})

		NvimInstance.RegisterHandler("MarkdownCodeBlocks", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 2 {
					continue
				}
				bufNr := utils.ReflectToInt(data[0])
				blocksRaw, ok := data[1].([]interface{})
				if !ok {
					continue
				}
				blocks := parseMarkdownCodeBlocks(blocksRaw)
				utils.Log(fmt.Sprintf("MarkdownCodeBlocks: bufNr=%d blocks=%v", bufNr, blocks))
				NvimScreen.setCodeBlockMetadata(bufNr, blocks)
			}
		})

		NvimInstance.RegisterHandler("MarkdownInlineCode", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 2 {
					continue
				}
				bufNr := utils.ReflectToInt(data[0])
				codesRaw, ok := data[1].([]interface{})
				if !ok {
					continue
				}
				codes := parseMarkdownInlineCode(codesRaw)
				NvimScreen.setInlineCodeMetadata(bufNr, codes)
			}
		})

		NvimInstance.RegisterHandler("CompletionShow", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 3 {
					continue
				}
				itemsRaw, ok := data[0].([]interface{})
				if !ok {
					continue
				}
				selectedIdx := utils.ReflectToInt(data[1])
				col := utils.ReflectToInt(data[2])

				items := parseCompletionItems(itemsRaw)
				state := CompletionState{
					Items:         items,
					SelectedIndex: selectedIdx,
					Col:           col,
				}
				if App != nil {
					App.Event.Emit("completion-show", state)
				}
			}
		})

		NvimInstance.RegisterHandler("CompletionHide", func(updates ...[]interface{}) {
			if App != nil {
				App.Event.Emit("completion-hide", struct{}{})
			}
		})

		NvimInstance.RegisterHandler("CompletionSelect", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 1 {
					continue
				}
				idx := utils.ReflectToInt(data[0])
				if App != nil {
					App.Event.Emit("completion-select", idx)
				}
			}
		})

		NvimInstance.RegisterHandler("CompletionDocumentation", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 3 {
					utils.Log("CompletionDocumentation: insufficient data")
					continue
				}
				docText, _ := data[0].(string)
				docKind, _ := data[1].(string)
				detail, _ := data[2].(string)

				doc := CompletionDocumentation{
					Text:   docText,
					Kind:   docKind,
					Detail: detail,
				}


				if App != nil {
					App.Event.Emit("completion-documentation", doc)
				}
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

		// Read colorcolumn and cursorline settings before disabling
		readColorColumn(NvimScreen)
		readCursorLine(NvimScreen)

		// Disable neovim's gutter, wrapping, colorcolumn, and cursorline — we render these ourselves.
		NvimInstance.Command("set nonumber norelativenumber signcolumn=no foldcolumn=0 nowrap colorcolumn= nocursorline conceallevel=0")

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

		// Set up completion bridge for blink.cmp
		if err := NvimInstance.ExecLua(completionLua, nil, NvimInstance.ChannelID()); err != nil {
			utils.Log(fmt.Sprintf("Error loading completion Lua: %v", err))
		}

		// Override split commands to create external windows (OS-level splits)
		if err := NvimInstance.ExecLua(externalSplitsLua, nil); err != nil {
			utils.Log(fmt.Sprintf("Error loading external splits Lua: %v", err))
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
		if App != nil {
			App.Quit()
		}
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

func readCursorLine(screen *Screen) {
	var enabled bool
	err := NvimInstance.ExecLua("return vim.opt.cursorline:get()", &enabled)
	if err != nil {
		utils.Log(fmt.Sprintf("Error reading cursorline: %v", err))
		return
	}
	screen.CursorLineEnabled = enabled

	if !enabled {
		return
	}

	var hlResult map[string]interface{}
	err = NvimInstance.ExecLua("return vim.api.nvim_get_hl(0, {name='CursorLine'})", &hlResult)
	if err != nil {
		utils.Log(fmt.Sprintf("Error reading CursorLine highlight: %v", err))
		return
	}

	if bg, ok := hlResult["bg"]; ok {
		bgInt := utils.ReflectToInt(bg)
		screen.CursorLineColor = fmt.Sprintf("#%06x", bgInt)
	}
	utils.Log(fmt.Sprintf("CursorLine: enabled=%v color=%s", screen.CursorLineEnabled, screen.CursorLineColor))
}

func isVisualMode(mode string) bool {
	return mode == "v" || mode == "V" || mode == "\x16" // Normal, line, and block visual modes
}
