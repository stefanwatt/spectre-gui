package main

import (
	"context"
	"fmt"
	"time"

	ext "spectre-gui/external-tools"
	filewatcher "spectre-gui/file-watcher"
	"spectre-gui/match"
	utils "spectre-gui/utils"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) Search(
	search_term string,
	dir string,
	include string,
	exclude string,
	flags []string,
	replace_term string,
	preserve_case bool,
) []match.SearchResult {
	utils.Log(fmt.Sprintf("searching...\nsearch_term: %s\ndir: %s\ninclude: %s\nexclude: %s\nflags: %s\nreplace_term:%s\npreserve_case:%v", search_term, dir, include, exclude, flags, replace_term, preserve_case))
	if search_term == "" {
		return []match.SearchResult{}
	}
	ctx, update_dir := filewatcher.InitContext(a.dir, dir, a.ctx)
	a.dir = update_dir
	rg_lines, err := ext.Ripgrep(
		search_term,
		replace_term,
		dir,
		include,
		exclude,
		has_flag("case-sensitive", flags),
		has_flag("regex", flags),
		has_flag("match_whole_word", flags),
		preserve_case,
	)
	if err != nil {
		utils.Log(fmt.Sprintf("ripgrep error: %s", err))
		return []match.SearchResult{}
	}
	use_regex := has_flag("regex", flags)
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
			use_regex,
		)
	})
	a.current_matches = matches
	search_results := match.MapSearchResult(matches)
	dirs := match.MapDirs(search_results)
	// TODO: how to handle errors in a go routine?
	go filewatcher.WatchFiles(ctx, dirs, dir, on_write, on_delete)
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
	ext.Sed(
		replaced_match.Row,
		replaced_match.Col,
		replaced_match.AbsolutePath,
		search_term,
		replace_term,
		preserve_case,
	)
	a.current_matches = utils.Filter(a.current_matches, func(m match.Match) bool {
		return m.FileName != replaced_match.FileName || m.Row != replaced_match.Row || m.Col != replaced_match.Col
	})
}

func (a *App) ReplaceAll(
	search_term string,
	replace_term string,
	dir string,
	include string,
	exclude string,
	flags []string,
	preserve_case bool,
) {
	utils.Log(fmt.Sprintf("replacing all: \nsearch_term: %s\nreplace_term: %s", search_term, replace_term))
	utils.Log("calling sed")
	rg_lines, err := ext.Ripgrep(
		search_term,
		replace_term,
		dir,
		include,
		exclude,
		has_flag("case-sensitive", flags),
		has_flag("regex", flags),
		has_flag("match_whole_word", flags),
		preserve_case,
	)
	if err != nil {
		utils.Log(fmt.Sprintf("ripgrep error: %s", err))
		return
	}
	use_regex := has_flag("regex", flags)
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
			use_regex,
		)
	})
	for _, match := range matches {
		ext.Sed(match.Row, match.Col, match.AbsolutePath, search_term, replace_term, preserve_case)
	}
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
		Runtime.EventsEmit(ctx, "files-changed")
	})
}

func on_delete(event fsnotify.Event, ctx context.Context) {
	path := event.Name
	if path[len(path)-1:] == "~" {
		return
	}

	utils.Log("on_delete")
	debounced_files_changed(func() {
		Runtime.EventsEmit(ctx, "files-changed")
	})
}

func has_flag(flag string, flags []string) bool {
	_, err := utils.Find(flags, func(f string) bool {
		return f == flag
	})
	return err == nil
}
