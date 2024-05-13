package main

import (
	"context"
	"fmt"
	"time"

	ext "spectre-gui/external-tools"
	filewatcher "spectre-gui/file-watcher"
	"spectre-gui/highlighting"
	"spectre-gui/match"
	"spectre-gui/undo"
	"spectre-gui/utils"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	undo_stack    = undo.UndoStack{}
	INFO_LEVEL    = "info"
	SUCCESS_LEVEL = "success"
	WARNING_LEVEL = "warning"
	ERROR_LEVEL   = "error"
	DELETE        = "file-deleted"
	REPLACE       = "file-replaced"
	REPLACE_ALL   = "replaced-all"
	UNDO          = "undo"
	TOAST         = "toast"
	write_event   = REPLACE
)

func (a *App) GetAppState() AppState {
	return a.State
}

type RglineRgInfoLexer struct {
	rg_line string
	rg_info ext.RipgrepInfo
	lexer   chroma.Lexer
}

func (a *App) Search(
	search_term string,
	replace_term string,
	dir string,
	exclude string,
	include string,
	case_sensitive bool,
	regex bool,
	match_whole_word bool,
	preserve_case bool,
) []match.SearchResult {
	utils.Log(fmt.Sprintf("searching...\nsearch_term: %s\ndir: %s\ninclude: %s\nexclude: %s\nreplace_term:%s\npreserve_case:%v", search_term, dir, include, exclude, replace_term, preserve_case))
	utils.LogTime("Search")
	if search_term == "" {
		return []match.SearchResult{}
	}
	utils.StartTime = time.Now()
	if a.search_ctx.cancel_func != nil {
		a.search_ctx.cancel_func()
	}
	var search_ctx context.Context
	search_ctx, a.search_ctx.cancel_func = context.WithCancel(context.Background())
	ctx, update_dir := filewatcher.InitContext(a.State.Dir, dir, a.ctx)
	a.State.Dir = update_dir
	utils.LogTimeSinceLast("before ripgrep")
	rg_lines, err := ext.Ripgrep(
		search_ctx,
		search_term,
		replace_term,
		dir,
		include,
		exclude,
		case_sensitive,
		regex,
		match_whole_word,
		preserve_case,
	)
	utils.LogTimeSinceLast("after ripgrep")
	if err != nil {
		utils.Log(fmt.Sprintf("ripgrep error: %s", err))
		if ctx.Err() == context.Canceled {
			utils.Log("Search was canceled")
		}
		return match.MapSearchResult(a.State.CurrentMatches)
	}
	// if len(rg_lines) > 0 {
	// 	rg_lines = rg_lines[:1]
	// }
	triples := utils.MapArray(rg_lines, func(line string) RglineRgInfoLexer {
		rg_info := ext.MapRipgrepInfo(line)
		lexer := highlighting.MatchLexer(rg_info.Path)
		return RglineRgInfoLexer{
			rg_line: line,
			rg_info: rg_info,
			lexer:   lexer,
		}
	})

	matches := utils.MapArrayConcurrent(triples, func(triple RglineRgInfoLexer) match.Match {
		line := triple.rg_line
		rg_info := triple.rg_info
		lexer := triple.lexer
		m := match.MapMatch(
			line,
			rg_info.Path,
			rg_info.MatchedText,
			rg_info.Row,
			rg_info.Col,
			search_term,
			replace_term,
			preserve_case,
			regex,
			lexer,
		)
		return m
	})
	utils.LogTimeSinceLast("map array")
	a.State.CurrentMatches = matches
	search_results := match.MapSearchResult(matches)
	dirs := match.MapDirs(search_results)
	// TODO: how to handle errors in a go routine?
	go filewatcher.WatchFiles(ctx, dirs, dir, on_write, on_delete)
	utils.LogTime("search")
	return search_results
}

func (a *App) Replace(
	replaced_match match.Match,
	search_term string,
	replace_term string,
	preserve_case bool,
) {
	utils.Log(fmt.Sprintf(
		"replacing in file: %s\nmatched line: %s\nsearch_term: %s\nreplace_term: %s\npreserve_case: %v",
		replaced_match.FileName,
		replaced_match.MatchedLine,
		search_term,
		replace_term,
		preserve_case,
	))
	utils.Log("calling sed")
	write_event = REPLACE
	ext.Replace(
		replaced_match.Row,
		replaced_match.Col,
		replaced_match.AbsolutePath,
		replaced_match.MatchedText,
		replaced_match.ReplacementText,
	)
	a.State.CurrentMatches = utils.Filter(a.State.CurrentMatches, func(m match.Match) bool {
		return m.FileName != replaced_match.FileName || m.Row != replaced_match.Row || m.Col != replaced_match.Col
	})
	replace_op := undo.ReplaceOp{
		Path:         replaced_match.AbsolutePath,
		Row:          replaced_match.Row,
		OriginalText: replaced_match.MatchedLine,
	}
	replace_actions := []undo.ReplaceOp{replace_op}
	undo_stack.Push(undo.ReplaceAction{
		Actions: replace_actions,
	})
	spawn_toast(a.ctx, SUCCESS_LEVEL, "Match replaced")
}

func (a *App) ReplaceAll(
	search_term string,
	replace_term string,
	dir string,
	exclude string,
	include string,
	case_sensitive bool,
	regex bool,
	match_whole_word bool,
	preserve_case bool,
) {
	utils.Log(fmt.Sprintf("replacing all: \nsearch_term: %s\nreplace_term: %s", search_term, replace_term))
	utils.Log("calling sed")

	if a.search_ctx.cancel_func != nil {
		a.search_ctx.cancel_func()
	}
	var search_ctx context.Context
	search_ctx, a.search_ctx.cancel_func = context.WithCancel(context.Background())
	rg_lines, err := ext.Ripgrep(
		search_ctx,
		search_term,
		replace_term,
		dir,
		include,
		exclude,
		case_sensitive,
		regex,
		match_whole_word,
		preserve_case,
	)
	if err != nil {
		utils.Log(fmt.Sprintf("ripgrep error: %s", err))
		return
	}
	lexer := lexers.Get("plaintext")
	matches := utils.MapArray(rg_lines, func(line string) match.Match {
		rg_info := ext.MapRipgrepInfo(line)
		return match.MapMatch(
			line,
			rg_info.Path,
			rg_info.MatchedText,
			rg_info.Row,
			rg_info.Col,
			search_term,
			replace_term,
			preserve_case,
			regex,
			lexer,
		)
	})
	write_event = REPLACE_ALL
	for _, match := range matches {
		ext.Replace(
			match.Row,
			match.Col,
			match.AbsolutePath,
			match.MatchedText,
			match.ReplacementText,
		)
	}
	replace_actions := utils.MapArray(matches, func(match match.Match) undo.ReplaceOp {
		return undo.ReplaceOp{
			Path:         match.AbsolutePath,
			Row:          match.Row,
			OriginalText: match.MatchedLine,
		}
	})
	undo_stack.Push(undo.ReplaceAction{
		Actions: replace_actions,
	})
	spawn_toast(a.ctx, SUCCESS_LEVEL, "Replaced all matches")
}

func (a *App) Undo() {
	if undo_stack.IsEmpty() {
		utils.Log("undo stack is empty - cannot pop")
		spawn_toast(a.ctx, INFO_LEVEL, "Nothing to undo")
		return
	}
	undo_action := undo_stack.Pop()
	write_event = UNDO
	for _, action := range undo_action.Actions {
		ext.ReplaceLine(action.Path, action.Row, action.OriginalText)
	}
	spawn_toast(a.ctx, SUCCESS_LEVEL, "Reverted last action")
}

var debounced_files_changed = debounce.New(100 * time.Millisecond)

var start time.Time

func on_write(event fsnotify.Event, ctx context.Context) {
	path := event.Name
	if path[len(path)-1:] == "~" {
		return
	}
	utils.Log("on_write")
	debounced_files_changed(func() {
		Runtime.EventsEmit(ctx, write_event)
	})
}

func on_delete(event fsnotify.Event, ctx context.Context) {
	path := event.Name
	if path[len(path)-1:] == "~" {
		return
	}

	utils.Log("on_delete")
	debounced_files_changed(func() {
		Runtime.EventsEmit(ctx, DELETE)
	})
}

func has_flag(flag string, flags []string) bool {
	_, err := utils.Find(flags, func(f string) bool {
		return f == flag
	})
	return err == nil
}

func spawn_toast(ctx context.Context, level string, message string) {
	Runtime.EventsEmit(ctx, TOAST, level, message)
}
