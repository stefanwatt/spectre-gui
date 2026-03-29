package neovim

import (
	"context"
	"fmt"
	"nvim-gui/features"
	"nvim-gui/utils"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/neovim/go-client/nvim"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	App         *application.App
	currentMode string
	NvimClient  *nvim.Nvim
)

func SetApp(app *application.App) {
	App = app
	SetEventEmitter(wailsEventEmitter{app: app})
}

var NvimScreen *Screen

func CalculateGridSize(windowWidth, windowHeight int) (rows, cols int) {
	cellWidth := 12
	cellHeight := 28
	statusLineHeight := 36
	availableHeight := windowHeight - statusLineHeight
	cols = windowWidth / cellWidth
	rows = availableHeight / cellHeight
	log.Debug(fmt.Sprintf("CalculateGridSize rows=%d cols=%d windowHeight=%d windowWidth=%d", rows, cols, windowHeight, windowWidth))
	return rows, cols
}

func StartListening(ctx context.Context) {
	var err error
	// Use default grid size for initial creation
	// The frontend calls OnResize() once mounted to set correct dimensions
	rows, cols := 24, 80
	log.Debug(fmt.Sprintf("StartListening initializing screen with default rows=%d cols=%d", rows, cols))
	NvimScreen = NewScreen(ctx, cols, rows, App)

	// Note: resize and request-state handlers are now service methods on App:
	// - App.OnResize(width, height) called from frontend
	// - App.RequestState() called from frontend

	nvimCtx, nvimCancel := context.WithCancel(ctx)
	nvimExitChan := make(chan struct{})

	nvimArgs := []string{"--embed", "+\"set nonumber\"", "+\"set norelativenumber\""}
	if len(os.Args) > 1 {
		nvimArgs = append(nvimArgs, os.Args[1])
	}

	NvimClient, err = nvim.NewChildProcess(
		nvim.ChildProcessCommand("nvim"),
		nvim.ChildProcessArgs(nvimArgs...),
		nvim.ChildProcessContext(nvimCtx),
	)
	NvimClient.SetVar("nvim_gui", true)
	NvimClient.SetVar("nvim_gui_channel", NvimClient.ChannelID())
	NvimClient.SetVar("keymaps", keymaps)

	if err != nil {
		log.Error(err)
		nvimCancel()
		if App != nil {
			App.Quit()
		}
		return
	}

	// Run a goroutine to handle Neovim serving and exit
	go func() {
		defer close(nvimExitChan)
		defer NvimClient.Close()

		// Register all handlers BEFORE AttachUI so no events are missed.
		// The go-client's internal goroutine dispatches notifications immediately;
		// if handlers aren't registered when hl_attr_define arrives, highlights are lost.
		NvimClient.RegisterHandler("redraw", func(updates ...[]interface{}) {
			HandleRedraw(updates)
		})

		channelID := NvimClient.ChannelID()

		// The Lua script creates a User autocmd that matches the mini.files patterns.
		// When triggered, it fires vim.rpcnotify to send the data to your Go channel.
		luaScript := `
    local chan_id = ... -- The first argument passed to ExecLua

    vim.api.nvim_create_autocmd("User", {
        pattern = {
            "MiniFilesExplorerOpen",
            "MiniFilesExplorerClose",
            "MiniFilesWindowOpen",
            "MiniFilesWindowUpdate",
            "MiniFilesBufferCreate",
            "MiniFilesBufferUpdate"
        },
        callback = function(args)
            vim.rpcnotify(chan_id, "MiniFilesBridge", args.match, args.data or {})
        end
    })
`

		// Execute the Lua script, passing the channelID as the argument (...)
		var result interface{}
		err := NvimClient.ExecLua(luaScript, &result, channelID)
		if err != nil {
			log.Debug(fmt.Sprintf("Failed to setup autocmd bridge: %v", err))
		}
		err = EnsureMiniFilesPatched()
		if err != nil {
			log.Debug(fmt.Sprintf("[minifiles] failed to load patched mini.files: %v", err))
		}

		NvimClient.RegisterHandler("MiniFilesBridge", func(eventName string, data interface{}) {
			fileExplorerRegistry := GetFileExplorerRegistry()
			if fileExplorerRegistry == nil {
				log.Debug("[minifiles] MiniFilesBridge: file explorer registry is nil, ignoring event")
				return
			}
			switch eventName {
			case "MiniFilesExplorerOpen":
				fileExplorerRegistry.SetActive(true)
				// TODO: verify this is not needed anymore
				// NvimScreen.FileExplorer.CurrentWinMode = NvimScreen.Mode
			case "MiniFilesExplorerClose":
				fileExplorerRegistry.SetActive(false)
			case "MiniFilesBufferCreate":
				fileExplorerRegistry.SetActive(true)
				dataMap, ok := data.(map[string]interface{})
				if !ok {
					return
				}
				bufNr := utils.ReflectToInt(dataMap["buf_id"])
				column, _ := dataMap["column"].(string)
				switch column {
				case "parent":
					fileExplorerRegistry.UpdateParentPaneBufnr(bufNr)
				case "current":
					fileExplorerRegistry.UpdateCurrentPaneBufnr(bufNr)
				case "preview":
					fileExplorerRegistry.UpdatePreviewPaneBufnr(bufNr)
				}
			case "MiniFilesWindowOpen":
				fileExplorerRegistry.SetActive(true)
				dataMap, ok := data.(map[string]interface{})
				if !ok {
					return
				}
				winId := utils.ReflectToInt(dataMap["win_id"])
				bufNr := utils.ReflectToInt(dataMap["buf_id"])
				column, _ := dataMap["column"].(string)
				switch column {
				case "parent":
					fileExplorerRegistry.AssignParentPane(winId, bufNr)
				case "current":
					fileExplorerRegistry.AssignCurrentPane(winId, bufNr)
				case "preview":
					fileExplorerRegistry.AssignPreviewPane(winId, bufNr)
					NvimScreen.applyFileExplorerPreviewSize()
				}
			case "MiniFilesWindowUpdate":
				dataMap, ok := data.(map[string]interface{})
				if !ok {
					log.Debug("[minifiles] MiniFilesWindowUpdate: invalid data format")
					return
				}
				column, _ := dataMap["column"].(string)
				if column == "" {
					log.Debug("[minifiles] MiniFilesWindowUpdate: missing column")
					return
				}
				winId := utils.ReflectToInt(dataMap["win_id"])
				bufNr := utils.ReflectToInt(dataMap["buf_id"])
				switch column {
				case "parent":
					fileExplorerRegistry.AssignParentPane(winId, bufNr)
				case "current":
					fileExplorerRegistry.AssignCurrentPane(winId, bufNr)
				case "preview":
					fileExplorerRegistry.AssignPreviewPane(winId, bufNr)
					// TODO: what about preview HasPreviewTargetSize and previewTargetCols
					NvimScreen.applyFileExplorerPreviewSize()
				default:
					log.Debug(fmt.Sprintf("[minifiles] MiniFilesWindowUpdate: unsupported column=%s", column))
				}
			case "MiniFilesBufferUpdate":
				dataMap, ok := data.(map[string]interface{})
				if !ok {
					log.Debug("MiniFilesBufferUpdate: invalid data format")
					return
				}
				bufNr := utils.ReflectToInt(dataMap["buf_id"])
				if bufNr < 1 {
					return
				}
				lineMap, err := GetMiniFilesDirectoryLineMap(bufNr)
				if err != nil {
					return
				}
				fileExplorerRegistry.SetDirectoryLineMap(bufNr, lineMap)
			}
		})

		NvimClient.RegisterHandler("BufEnter", func(_ *nvim.Nvim, data []string) {
			assert(len(data) == 1, "BufEnter: malformed data")
			filepath := filepath.Base(data[0])
			EmitEvent("BufEnter", filepath)
		})

		NvimClient.RegisterHandler("MarkdownTables", func(updates ...[]interface{}) {
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
				log.Debug(fmt.Sprintf("MarkdownTables received: bufNr=%d, tables=%d", bufNr, len(tables)))
				NvimScreen.setTableMetadata(bufNr, tables)
			}
		})

		NvimClient.RegisterHandler("MarkdownImages", func(updates ...[]interface{}) {
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
				log.Debug(fmt.Sprintf("MarkdownImages received: bufNr=%d, images=%d", bufNr, len(images)))
				NvimScreen.setImageMetadata(bufNr, images)
			}
		})

		NvimClient.RegisterHandler("MarkdownHeadings", func(updates ...[]interface{}) {
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

		NvimClient.RegisterHandler("MarkdownTasks", func(updates ...[]interface{}) {
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

		NvimClient.RegisterHandler("MarkdownCodeBlocks", func(updates ...[]interface{}) {
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
				log.Debug(fmt.Sprintf("MarkdownCodeBlocks: bufNr=%d blocks=%v", bufNr, blocks))
				NvimScreen.setCodeBlockMetadata(bufNr, blocks)
			}
		})

		NvimClient.RegisterHandler("MarkdownInlineCode", func(updates ...[]interface{}) {
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

		NvimClient.RegisterHandler("CompletionShow", func(updates ...[]interface{}) {
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

				items := features.ParseCompletionItems(itemsRaw)
				state := features.CompletionState{
					Items:         items,
					SelectedIndex: selectedIdx,
					Col:           col,
				}
				EmitEvent("completion-show", state)
			}
		})

		NvimClient.RegisterHandler("CompletionHide", func(updates ...[]interface{}) {
			EmitEvent("completion-hide", struct{}{})
		})

		NvimClient.RegisterHandler("CompletionSelect", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 1 {
					continue
				}
				idx := utils.ReflectToInt(data[0])
				EmitEvent("completion-select", idx)
			}
		})

		NvimClient.RegisterHandler("CompletionDocumentation", func(updates ...[]interface{}) {
			for _, data := range updates {
				if len(data) < 3 {
					log.Debug("CompletionDocumentation: insufficient data")
					continue
				}
				docText, _ := data[0].(string)
				docKind, _ := data[1].(string)
				detail, _ := data[2].(string)

				doc := features.CompletionDocumentation{
					Text:   docText,
					Kind:   docKind,
					Detail: detail,
				}

				EmitEvent("completion-documentation", doc)
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

		err = NvimClient.AttachUI(cols, rows, opts)
		if err != nil {
			log.Error(err.Error())
			nvimCancel()
			return
		}

		// Read colorcolumn and cursorline settings before disabling
		readColorColumn(NvimScreen)
		readCursorLine(NvimScreen)

		// Disable neovim's gutter, wrapping, colorcolumn, and cursorline — we render these ourselves.
		NvimClient.Command("set nonumber norelativenumber signcolumn=no foldcolumn=0 nowrap colorcolumn= nocursorline conceallevel=0")
		// Use autocmds so these stay enforced:
		// 1. On every new/entered window (user config autocmds can re-enable them).
		// 2. On VimEnter, apply to ALL existing windows (handles startup with multiple splits).
		NvimClient.ExecLua(`
			local function disable_gutter()
				vim.wo.number = false
				vim.wo.relativenumber = false
				vim.wo.signcolumn = "no"
				vim.wo.foldcolumn = "0"
			end

			vim.api.nvim_create_autocmd({"WinNew", "WinEnter", "BufWinEnter"}, {
				callback = disable_gutter,
			})

			vim.api.nvim_create_autocmd("VimEnter", {
				once = true,
				callback = function()
					for _, win in ipairs(vim.api.nvim_list_wins()) do
						vim.api.nvim_win_call(win, disable_gutter)
					end
				end,
			})
		`, nil)

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
			log.Debug(fmt.Sprintf("Opening test file: %s", absPath))
			err := NvimClient.Command(fmt.Sprintf("edit %s", absPath))
			if err != nil {
				log.Debug(fmt.Sprintf("Error opening test file: %v", err))
			}
		}

		SetupKeymaps()

		// Set up markdown table detection via treesitter
		if err := NvimClient.ExecLua(markdownTablesLua, nil, NvimClient.ChannelID()); err != nil {
			log.Debug(fmt.Sprintf("Error loading markdown tables Lua: %v", err))
		}

		// Set up completion bridge for blink.cmp
		if err := NvimClient.ExecLua(features.CompletionLua, nil, NvimClient.ChannelID()); err != nil {
			log.Debug(fmt.Sprintf("Error loading completion Lua: %v", err))
		}

		if err := NvimClient.Serve(); err != nil {
			log.Debug(fmt.Sprintf("Neovim process terminated: %v\n%s", err, debug.Stack()))
		}

		// Neovim has exited, signal to quit the app
		log.Debug("Neovim process has terminated, quitting application")
	}()

	// Wait for Neovim to exit or context to be cancelled
	select {
	case <-nvimExitChan:
		// Neovim exited, quit the app
		log.Debug("Detected Neovim exit, shutting down application")
		if App != nil {
			App.Quit()
		}
	case <-ctx.Done():
		// Parent context was cancelled
		nvimCancel()
	}

	log.Debug("listening terminating")
}

func readColorColumn(screen *Screen) {
	// Read colorcolumn setting (returns a list of strings like {"80"} or {"+1", "120"})
	var ccValues []interface{}
	err := NvimClient.ExecLua("return vim.opt.colorcolumn:get()", &ccValues)
	if err != nil {
		log.Debug(fmt.Sprintf("Error reading colorcolumn: %v", err))
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
	err = NvimClient.ExecLua("return vim.api.nvim_get_hl(0, {name='ColorColumn'})", &hlResult)
	if err != nil {
		log.Debug(fmt.Sprintf("Error reading ColorColumn highlight: %v", err))
		return
	}

	if bg, ok := hlResult["bg"]; ok {
		bgInt := utils.ReflectToInt(bg)
		screen.ColorColumnColor = fmt.Sprintf("#%06x", bgInt)
	}
	log.Debug(fmt.Sprintf("ColorColumn: columns=%v color=%s", screen.ColorColumns, screen.ColorColumnColor))
}

func readCursorLine(screen *Screen) {
	var enabled bool
	err := NvimClient.ExecLua("return vim.opt.cursorline:get()", &enabled)
	if err != nil {
		log.Debug(fmt.Sprintf("Error reading cursorline: %v", err))
		return
	}
	screen.CursorLineEnabled = enabled

	if !enabled {
		return
	}

	var hlResult map[string]interface{}
	err = NvimClient.ExecLua("return vim.api.nvim_get_hl(0, {name='CursorLine'})", &hlResult)
	if err != nil {
		log.Debug(fmt.Sprintf("Error reading CursorLine highlight: %v", err))
		return
	}

	if bg, ok := hlResult["bg"]; ok {
		bgInt := utils.ReflectToInt(bg)
		screen.CursorLineColor = fmt.Sprintf("#%06x", bgInt)
	}
	log.Debug(fmt.Sprintf("CursorLine: enabled=%v color=%s", screen.CursorLineEnabled, screen.CursorLineColor))
}

func isVisualMode(mode string) bool {
	return mode == "v" || mode == "V" || mode == "\x16" // Normal, line, and block visual modes
}
