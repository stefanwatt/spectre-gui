package main

import (
	"context"

	"nvim-gui/neovim"
	"nvim-gui/utils"
)

var ctx context.Context

type App struct {
	ctx        context.Context
	Servername string
	File       string
}

func NewApp() *App {
	return &App{}
}

func (a *App) mounted(ctx context.Context) {
	if a.Servername != "" {
		go neovim.StartListening(a.Servername, a.ctx)
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	utils.SetupLog()
}
