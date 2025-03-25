package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/jessevdk/go-flags"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/build
var assets embed.FS

type Options struct {
	Servername string `short:"n" long:"servername" description:"neovim servername" required:"false"`
	File       string `short:"f" long:"filename" description:"file to open" required:"false"`
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

func spawnNeovim(servername string, filename string) (*exec.Cmd, error) {
	cmd := exec.Command("nvim", "--embed", "--listen", servername, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, cmd.Start()
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
	// app.Servername = opts.Servername
	app.Servername = "/tmp/nvimsocket"
	// app.File = opts.File
	app.File = "/tmp/foo.lua"
	err = deleteIfExists(app.Servername)
	if err != nil {
		fmt.Println("error deleting nvim socket")
	}
	cmd, err := spawnNeovim(app.Servername, app.File)
	if err != nil {
		panic(err)
	}
	cmd.Wait()

	err = wails.Run(&options.App{
		Title:              "nvim-gui",
		LogLevel:           logger.ERROR,
		LogLevelProduction: logger.ERROR,
		Width:              1024,
		Height:             768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
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
