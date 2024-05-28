package main

import (
	"context"
	"fmt"
	"time"

	ext "spectre-gui/external-tools"
	filewatcher "spectre-gui/file-watcher"
	"spectre-gui/highlighting"
	"spectre-gui/match"
	"spectre-gui/neovim"
	"spectre-gui/undo"
	"spectre-gui/utils"

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
	page_size     = 20
)

func (a *App) GetAppState() AppState {
	return a.State
}

type SearchResult struct {
	GroupedMatches []match.MatchesOfFile
	PageIndex      int
	TotalPages     int
	TotalResults   int
	TotalFiles     int
}

func (a *App) GetRoute() string {
	return a.Mode
}

func (a *App) OpenMatch(path string, row int, col int) {
	utils.Log("a.Servername")
	err := neovim.OpenFileAt(path, row, col, a.Servername)
	if err != nil {
		utils.Log(err.Error())
	}
}

func (a *App) AddMatchesToQuickfixList() {
}

func (a *App) SendKey(key string, ctrl bool, alt bool, shift bool) {
	neovim.SendKey(key, ctrl, alt, shift, a.Servername)
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
) SearchResult {
	if search_term == "" {
		return SearchResult{}
	}
	utils.StartTime = time.Now()
	if a.search_ctx.cancel_func != nil {
		a.search_ctx.cancel_func()
	}
	var search_ctx context.Context
	search_ctx, a.search_ctx.cancel_func = context.WithCancel(context.Background())
	ctx, update_dir := filewatcher.InitContext(a.State.Dir, dir, a.ctx)
	a.State.Dir = update_dir
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
		if ctx.Err() == context.Canceled {
			utils.Log("Search was canceled")
		}
		return SearchResult{}
	}
	a.State.Pagination = map_pagination(rg_lines)
	matches := utils.MapArrayConcurrent(a.State.Pagination.Pages[0].Matches, func(page_match PageMatch) match.Match {
		line := page_match.RgLine
		rg_info := ext.MapRipgrepInfo(line)
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
		)
		html, _ := highlighting.Highlight(
			m.MatchedLine,
			rg_info.Path,
			m.MatchedText,
			m.ReplacementText,
		)
		m.Html = html
		return m
	})
	grouped_matches := match.MapSearchResult(matches)
	dirs := match.MapDirs(grouped_matches)
	// TODO: how to handle errors in a go routine?
	go filewatcher.WatchFiles(ctx, dirs, dir, on_write, on_delete)

	paths := utils.MapArray(rg_lines, func(line string) string {
		return ext.MapRipgrepInfo(line).Path
	})
	total_files := utils.CountUniqueItems(paths)
	return SearchResult{
		GroupedMatches: grouped_matches,
		TotalPages:     len(a.State.Pagination.Pages),
		PageIndex:      a.State.Pagination.PageIndex,
		TotalResults:   len(rg_lines),
		TotalFiles:     total_files,
	}
}

func (a *App) GetPrevPage() SearchResult {
	if len(a.State.Pagination.Pages) == 0 {
		return SearchResult{}
	}
	a.State.Pagination.PageIndex--
	if a.State.Pagination.PageIndex < 0 {
		a.State.Pagination.PageIndex = len(a.State.Pagination.Pages) - 1
	}
	page := a.State.Pagination.Pages[a.State.Pagination.PageIndex]
	matches := utils.MapArrayConcurrent(page.Matches, func(page_match PageMatch) match.Match {
		line := page_match.RgLine
		rg_info := ext.MapRipgrepInfo(line)
		m := match.MapMatch(
			line,
			rg_info.Path,
			rg_info.MatchedText,
			rg_info.Row,
			rg_info.Col,
			a.State.SearchTerm,
			a.State.ReplaceTerm,
			a.State.PreserveCase,
			a.State.Regex,
		)
		html, _ := highlighting.Highlight(
			m.MatchedLine,
			rg_info.Path,
			m.MatchedText,
			m.ReplacementText,
		)
		m.Html = html
		return m
	})
	grouped_matches := match.MapSearchResult(matches)

	return SearchResult{
		GroupedMatches: grouped_matches,
		TotalPages:     len(a.State.Pagination.Pages),
		PageIndex:      a.State.Pagination.PageIndex,
		TotalResults:   a.State.TotalResults,
		TotalFiles:     a.State.TotalFiles,
	}
}

func (a *App) GetNextPage() SearchResult {
	if len(a.State.Pagination.Pages) == 0 {
		return SearchResult{}
	}
	a.State.Pagination.PageIndex++
	if a.State.Pagination.PageIndex >= len(a.State.Pagination.Pages) {
		a.State.Pagination.PageIndex = 0
	}
	page := a.State.Pagination.Pages[a.State.Pagination.PageIndex]
	matches := utils.MapArrayConcurrent(page.Matches, func(page_match PageMatch) match.Match {
		line := page_match.RgLine
		rg_info := ext.MapRipgrepInfo(line)
		m := match.MapMatch(
			line,
			rg_info.Path,
			rg_info.MatchedText,
			rg_info.Row,
			rg_info.Col,
			a.State.SearchTerm,
			a.State.ReplaceTerm,
			a.State.PreserveCase,
			a.State.Regex,
		)
		html, _ := highlighting.Highlight(
			m.MatchedLine,
			rg_info.Path,
			m.MatchedText,
			m.ReplacementText,
		)
		m.Html = html
		return m
	})
	grouped_matches := match.MapSearchResult(matches)

	return SearchResult{
		GroupedMatches: grouped_matches,
		TotalPages:     len(a.State.Pagination.Pages),
		PageIndex:      a.State.Pagination.PageIndex,
		TotalResults:   a.State.TotalResults,
		TotalFiles:     a.State.TotalFiles,
	}
}

func (a *App) Replace(
	replaced_match match.Match,
	search_term string,
	replace_term string,
	preserve_case bool,
) {
	write_event = REPLACE
	ext.Replace(
		replaced_match.Row,
		replaced_match.Col,
		replaced_match.AbsolutePath,
		replaced_match.MatchedText,
		replaced_match.ReplacementText,
	)
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

func (a *App) GetReplacementText(
	matched_line string,
	search_term string,
	replace_term string,
	use_regex bool,
) string {
	replacement_text, err := ext.GetReplacementText(matched_line, search_term, replace_term, use_regex)
	if err != nil {
		return replace_term
	}
	return replacement_text
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
	debounced_files_changed(func() {
		utils.Log("on_write")
		utils.Log(write_event)
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

func map_pagination(rg_lines []string) Pagination {
	chunks := utils.ChunkSlice(rg_lines, page_size)
	var pages []Page
	for i := 0; i < len(chunks); i++ {
		rg_lines := chunks[i]
		matches := utils.MapArray(rg_lines, func(line string) PageMatch {
			return PageMatch{
				RgLine: line,
				Match:  nil,
			}
		})
		page := Page{
			Index:   i,
			Matches: matches,
		}
		pages = append(pages, page)
	}
	return Pagination{
		PageIndex: 0,
		Pages:     pages,
	}
}
