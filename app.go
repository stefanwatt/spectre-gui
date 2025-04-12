package main

import (
	"context"

	"nvim-gui/neovim"
	"nvim-gui/picker"
	"nvim-gui/utils"
)

var ctx context.Context

type App struct {
	ctx  context.Context
	File string
	liveGrepPicker *picker.LiveGrepPicker
}

func NewApp() *App {
	return &App{
		liveGrepPicker: picker.NewLiveGrepPicker(),
	}
}

func (a *App) mounted(ctx context.Context) {
	go neovim.StartListening(a.ctx)
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	utils.SetupLog()
}
