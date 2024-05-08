package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var debounced_files_changed = debounce.New(100 * time.Millisecond)

var start time.Time

func on_write(event fsnotify.Event, ctx context.Context) {
	path := event.Name
	if path[len(path)-1:] == "~" {
		return
	}
	Log("on_write")
	debounced_files_changed(func() {
		Runtime.EventsEmit(ctx, "files-changed")
	})
}

func on_delete(event fsnotify.Event, ctx context.Context) {
	path := event.Name
	if path[len(path)-1:] == "~" {
		return
	}

	Log("on_delete")
	debounced_files_changed(func() {
		Runtime.EventsEmit(ctx, "files-changed")
	})
}

func (a *App) Search(search_term string, dir string, include string, exclude string, flags []string) RipgrepResult {
	Log(fmt.Sprintf("searching...\nsearch_term: %s\ndir: %s\ninclude: %s\nexclude: %s\nflags: %s", search_term, dir, include, exclude, flags))
	if a.dir != dir && a.close_dir_watcher != nil {
		a.close_dir_watcher()
		a.dir = dir
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.close_dir_watcher = cancel
	if search_term == "" {
		return RipgrepResult{}
	}
	matches, err := Ripgrep(search_term, dir, include, exclude, flags)
	if err != nil {
		Log(fmt.Sprintf("ripgrep error: %s", err))
		return RipgrepResult{}
	}

	a.current_matches = *matches
	grouped := group_by_property(*matches, func(match RipgrepMatch) string {
		return match.Path
	})
	go ObserveFiles(ctx, grouped, dir, on_write, on_delete)
	return grouped
}

func (a *App) Replace(replaced_match RipgrepMatch, search_term string, replace_term string, preserve_case bool) {
	Log(fmt.Sprintf("replacing in file: %s\nmatched line: %s\nsearch_term: %s\nreplace_term: %s\npreserve_case: %v", replaced_match.Path, replaced_match.MatchedLine, search_term, replace_term, preserve_case))
	Log("calling sed")
	Sed(replaced_match, search_term, replace_term)
	a.current_matches = Filter(a.current_matches, func(m RipgrepMatch) bool {
		return m.Path != replaced_match.Path || m.Row != replaced_match.Row || m.Col != replaced_match.Col
	})
}

func (a *App) ReplaceAll(search_term string, replace_term string, dir string, include string, exclude string, flags []string) {
	Log(fmt.Sprintf("replacing all: \nsearch_term: %s\nreplace_term: %s", search_term, replace_term))
	Log("calling sed")
	matches, err := Ripgrep(search_term, dir, include, exclude, flags)
	if err != nil {
		Log(fmt.Sprintf("ripgrep error: %s", err))
		return
	}
	for _, match := range *matches {
		Sed(match, search_term, replace_term)
	}
}

func group_by_property[T any, K comparable](items []T, getProperty func(T) K) map[K][]T {
	grouped := make(map[K][]T)

	for _, item := range items {
		key := getProperty(item)
		grouped[key] = append(grouped[key], item)
	}

	return grouped
}
