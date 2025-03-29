package main

import (
	"context"

	"nvim-gui/neovim"
	"nvim-gui/utils"
)

var ctx context.Context

type App struct {
	ctx        context.Context
	File       string
}

func NewApp() *App {
	return &App{}
}

func (a *App) mounted(ctx context.Context) {
		go neovim.StartListening( a.ctx)
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	utils.SetupLog()
}
