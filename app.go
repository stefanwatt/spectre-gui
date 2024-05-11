package main

import (
	"context"

	match "spectre-gui/match"
)

var ctx context.Context

type App struct {
	ctx             context.Context
	search_term     string
	replace_term    string
	dir             string
	include         string
	exclude         string
	current_matches []match.Match
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
