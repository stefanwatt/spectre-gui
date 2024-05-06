package main

import (
	"context"
)

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
