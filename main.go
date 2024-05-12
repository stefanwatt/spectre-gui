package main

import (
	"embed"
	"fmt"
	"os"

	"spectre-gui/lua"
	"spectre-gui/utils"

	"github.com/jessevdk/go-flags"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/build
var assets embed.FS

type Options struct {
	SearchTerm  string `short:"s" long:"search-term" description:"search term" required:"false"`
	ReplaceTerm string `short:"r" long:"replace-term" description:"replace term" required:"false"`
	Dir         string `short:"d" long:"dir" description:"Directory to search in" required:"false"`
	Include     string `short:"i" long:"include" description:"glob pattern eg.: */**.go to include in search" required:"false"`
	Exclude     string `short:"x" long:"exclude" description:"glob pattern eg.: */**.go to exclude from search" required:"false"`
}

func main() {
	app := NewApp()
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		// Handle error
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	config := lua.LoadConfig()
	utils.Log("LUA CONFIG")
	fmt.Println("config.CaseSensitive", config.CaseSensitive)
	fmt.Println("config.Regex", config.Regex)
	fmt.Println("config.MatchWholeWord", config.MatchWholeWord)
	fmt.Println("config.PreserveCase", config.PreserveCase)
	state := AppState{
		SearchTerm:     opts.SearchTerm,
		ReplaceTerm:    opts.ReplaceTerm,
		Dir:            opts.Dir,
		Include:        opts.Include,
		Exclude:        opts.Exclude,
		CaseSensitive:  config.CaseSensitive,
		Regex:          config.Regex,
		MatchWholeWord: config.MatchWholeWord,
		PreserveCase:   config.PreserveCase,
	}
	app.State = state
	err = wails.Run(&options.App{
		Title:              "spectre-gui",
		LogLevel:           logger.ERROR,
		LogLevelProduction: logger.ERROR,
		Width:              1024,
		Height:             768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
