package main

import (
	"context"

	"spectre-gui/match"
	"spectre-gui/utils"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var ctx context.Context

type AppState struct {
	SearchTerm     string
	ReplaceTerm    string
	Dir            string
	Include        string
	Exclude        string
	CaseSensitive  bool
	Regex          bool
	MatchWholeWord bool
	PreserveCase   bool
	Pagination     Pagination
	TotalResults   int
	TotalFiles     int
}

type PageMatch struct {
	RgLine string
	Match  *match.Match
}

type Page struct {
	Index   int
	Matches []PageMatch
}
type Pagination struct {
	PageIndex int
	Pages     []Page
}
type SearchContext struct {
	ctx         context.Context
	cancel_func context.CancelFunc
}

type App struct {
	ctx        context.Context
	State      AppState
	search_ctx SearchContext
	Mode       string
	Servername string
}

func NewApp() *App {
	return &App{}
}

func (a *App) mounted(ctx context.Context) {
	utils.Log("mounted mode: ", a.Mode)
	Runtime.EventsEmit(a.ctx, "change-url", a.Mode)
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	search_ctx, cancel := context.WithCancel(context.Background())
	a.search_ctx = SearchContext{
		ctx:         search_ctx,
		cancel_func: cancel,
	}
}
