package main

import (
	"context"
	"fmt"

	ext "spectre-gui/external-tools"
	"spectre-gui/match"
	utils "spectre-gui/utils"
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
	if a.dir != dir && a.close_dir_watcher != nil {
		a.close_dir_watcher()
		a.dir = dir
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.close_dir_watcher = cancel
	if search_term == "" {
		return []match.SearchResult{}
	}
	rg_lines, err := ext.Ripgrep(
		search_term,
		dir,
		include,
		exclude,
		flags,
		replace_term,
		preserve_case,
	)
	if err != nil {
		utils.Log(fmt.Sprintf("ripgrep error: %s", err))
		return []match.SearchResult{}
	}
	_, err = utils.Find(flags, func(flag string) bool {
		return flag == "regex"
	})
	use_regex := err == nil
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
	go utils.ObserveFiles(ctx, dirs, dir, utils.OnWrite, utils.OnDelete)
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
		dir,
		include,
		exclude,
		flags,
		replace_term,
		preserve_case,
	)
	if err != nil {
		utils.Log(fmt.Sprintf("ripgrep error: %s", err))
		return
	}
	_, err = utils.Find(flags, func(flag string) bool {
		return flag == "regex"
	})
	use_regex := err == nil
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
