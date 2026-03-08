package main

import (
	"embed"
	"fmt"
	"nvim-gui/neovim"
	"nvim-gui/picker"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/build
var assets embed.FS

type Options struct {
	File string `short:"f" long:"filename" description:"file to open" required:"false"`
}

func deleteIfExists(path string) error {
	var err error
	if _, err = os.Stat(path); err == nil {
		return os.Remove(path)
	} else if os.IsNotExist(err) {
		return nil
	}
	return err
}

func init() {
	// Register all events used in the application
	// Rendering & Layout events
	application.RegisterEvent[interface{}]("content-updated") // (winId int, tokens []Token)
	application.RegisterEvent[string]("highlight-css")
	application.RegisterEvent[interface{}]("viewport_changed") // map[string]interface{}
	application.RegisterEvent[interface{}]("window_opened")

	// Window Management events
	application.RegisterEvent[int]("hide-window")
	application.RegisterEvent[interface{}]("floating_windows") // []map[string]interface{}
	application.RegisterEvent[interface{}]("preview-window")   // map[string]interface{}
	application.RegisterEvent[int]("floating_window_closed")
	application.RegisterEvent[int]("preview-window-closed")

	// Cursor & Mode events
	application.RegisterEvent[interface{}]("cursor-changed") // CursorMoveEvent struct
	application.RegisterEvent[string]("mode-changed")

	// Command Line events
	application.RegisterEvent[interface{}]("cmdline_show") // map[string]interface{}
	application.RegisterEvent[interface{}]("cmdline_pos")  // map[string]interface{}
	application.RegisterEvent[interface{}]("cmdline_hide")

	// Completion events
	application.RegisterEvent[interface{}]("completion-show") // CompletionState
	application.RegisterEvent[interface{}]("completion-hide")
	application.RegisterEvent[int]("completion-select")
	application.RegisterEvent[interface{}]("completion-documentation") // CompletionDocumentation

	// Buffer events
	application.RegisterEvent[string]("BufEnter")

	// Picker Activation events
	application.RegisterEvent[interface{}]("show_live_grep")
	application.RegisterEvent[interface{}]("show-find-files")
	application.RegisterEvent[interface{}]("show-find-references")
	application.RegisterEvent[interface{}]("show-find-buffer-symbols")
	application.RegisterEvent[interface{}]("show-find-help")
	application.RegisterEvent[interface{}]("hide-live-rep")

	// File System events
	application.RegisterEvent[interface{}]("file-replaced")
	application.RegisterEvent[interface{}]("file-deleted")
	application.RegisterEvent[interface{}]("toast") // (level, message)
}

func main() {
	// WebKitGTK on Wayland uses DMABuf for hardware-accelerated rendering, but the
	// DMABuf surface doesn't resize correctly when the compositor resizes the window,
	// leaving transparent gaps. Disabling it forces WebKit to use a software surface
	// that resizes properly.
	os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")

	// Parse command line options
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Delete old log file
	err = deleteIfExists("/tmp/nvim-gui.log")
	if err != nil {
		fmt.Println("error deleting nvim socket")
	}

	// Create services
	appService := NewApp()
	pickerService := picker.NewPicker()
	liveGrepPicker := picker.NewLiveGrepPicker()
	referencesPicker := picker.NewReferencesPicker()
	symbolsPicker := picker.NewSymbolsPicker()

	// Create application
	app := application.New(application.Options{
		Name:        "nvim-gui",
		Description: "Neovim GUI with Wails + Svelte",
		Services: []application.Service{
			application.NewService(appService),
			application.NewService(pickerService),
			application.NewService(liveGrepPicker),
			application.NewService(referencesPicker),
			application.NewService(symbolsPicker),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	// Store app reference in services that need it
	appService.App = app
	pickerService.App = app
	liveGrepPicker.App = app

	// Pass app to neovim package for event emission
	neovim.SetApp(app)
	neovim.InitOSWindowManager(app)

	// Create the initial window — it will be reused by OSWindowManager
	// for the first neovim window (navigated to /window/{winId}?gridId={gridId})
	initialWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "nvim-gui",
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(39, 42, 56),
		URL:              "/",
	})
	neovim.SetInitialWindow(initialWindow)

	// Run application
	err = app.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
