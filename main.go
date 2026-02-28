package main

import (
	"context"
	"embed"
	"fmt"
	"nvim-gui/picker"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
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

func main() {
	app := NewApp()
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// app.File = opts.File
	err = deleteIfExists("/tmp/nvim-gui.log")
	if err != nil {
		fmt.Println("error deleting nvim socket")
	}

	liveGrepPicker := picker.NewLiveGrepPicker()
	referencesPicker := picker.NewReferencesPicker()
	symbolsPicker := picker.NewSymbolsPicker()
	picker := picker.NewPicker()

	startup := func(ctx context.Context) {
		app.startup(ctx)
		liveGrepPicker.Ctx = ctx
		picker.Ctx = ctx
	}

	err = wails.Run(&options.App{
		Title:              "nvim-gui",
		LogLevel:           logger.ERROR,
		LogLevelProduction: logger.ERROR,
		Width:              1024,
		Height:             768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Linux: &linux.Options{
			WebviewGpuPolicy: linux.WebviewGpuPolicyAlways,
		},
		BackgroundColour: &options.RGBA{R: 39, G: 42, B: 56, A: 1},
		OnStartup:        startup,
		OnDomReady:       app.mounted,
		Bind: []interface{}{
			app,
			liveGrepPicker,
			referencesPicker,
			symbolsPicker,
			picker,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
