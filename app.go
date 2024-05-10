package main

import (
	"context"

	match "spectre-gui/match"
)

var ctx context.Context

type App struct {
	ctx             context.Context
	dir             string
	current_matches []match.Match
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
