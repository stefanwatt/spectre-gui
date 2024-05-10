package main

import (
	"context"

	match "spectre-gui/match"
)

var ctx context.Context

type App struct {
	ctx               context.Context
	close_dir_watcher func()
	dir               string
	current_matches   []match.Match
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
