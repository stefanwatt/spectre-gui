package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var ctx context.Context

type App struct {
	ctx               context.Context
	close_dir_watcher func()
	dir               string
	current_matches   []RipgrepMatch
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func Log(text string) {
	message := "\n" + lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(text) + "\n"
	if ctx == nil {
		fmt.Println(message)
		return
	}
	Runtime.LogPrint(
		ctx,
		message,
	)
}
