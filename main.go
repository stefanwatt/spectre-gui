package main

import (
	"embed"
	"fmt"
	"net/http"
	"nvim-gui/utils"
	"os"
	"strings"

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

type FileLoader struct {
	http.Handler
}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (h *FileLoader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var err error
	requestedFilename := strings.TrimPrefix(req.URL.Path, "/")
	utils.Log("Requesting file:", requestedFilename)
	if requestedFilename == "nvim-hl.css" {
		requestedFilename = "/home/stefan/.config/nvim-gui/nvim-hl.css"
	}
	fileData, err := os.ReadFile(requestedFilename)
	if err != nil {
		utils.Log("couldnt get file: " + requestedFilename + "\nerror:" + err.Error())
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(fmt.Sprintf("Could not load file %s", requestedFilename)))
	}

	res.Write(fileData)
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

	err = wails.Run(&options.App{
		Title:              "nvim-gui",
		LogLevel:           logger.ERROR,
		LogLevelProduction: logger.ERROR,
		Width:              1024,
		Height:             768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewFileLoader(),
		},
		Linux: &linux.Options{
			WebviewGpuPolicy: linux.WebviewGpuPolicyAlways,
		},
		BackgroundColour: &options.RGBA{R: 39, G: 42, B: 56, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.mounted,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
