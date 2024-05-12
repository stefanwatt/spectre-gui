package main

import (
	"context"

	match "spectre-gui/match"
)

var ctx context.Context

type AppState struct {
	SearchTerm     string
	ReplaceTerm    string
	Dir            string
	Include        string
	Exclude        string
	CurrentMatches []match.Match
	CaseSensitive  bool
	Regex          bool
	MatchWholeWord bool
	PreserveCase   bool
}

type App struct {
	ctx   context.Context
	State AppState
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
