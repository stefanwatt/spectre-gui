package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher

func ObserveFiles(ctx context.Context,
	match_groups map[string][]RipgrepMatch,
	dir string,
	on_write func(fsnotify.Event, context.Context),
	on_delete func(fsnotify.Event, context.Context),
) {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		Log(err.Error())
		panic(err)
	}
	debouncedWrite := debounce.New(2000 * time.Millisecond)
	debouncedDelete := debounce.New(2000 * time.Millisecond)
	defer watcher.Close()
	err = watch_files(match_groups, dir)
	if err != nil {
		Log(err.Error())
		panic(err)
	}
	err = watcher.Add(dir)
	if err != nil {
		Log("skipping directory " + dir + " because of error")
		Log(err.Error())
		panic(err)
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write != 0 || event.Op&fsnotify.Create != 0 {
				debouncedWrite(func() {
					on_write(event, ctx)
				})
			}
			if event.Op&fsnotify.Remove != 0 {
				debouncedDelete(func() {
					on_delete(event, ctx)
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			Log(err.Error())
		}
	}
}

func watch_files(match_groups map[string][]RipgrepMatch, dir string) error {
	var dirs []string
	for path := range match_groups {
		current_dir := filepath.Dir(match_groups[path][0].AbsolutePath)
		_, err := Find(dirs, func(dir string) bool {
			return dir == current_dir
		})
		if err != nil {
			dirs = append(dirs, current_dir)
		}
	}

	for _, dir := range dirs {
		Log("Adding watcher for " + dir + "\n")
		filepath.Dir(dir)
		err := watcher.Add(dir)
		if err != nil {
			return err
		}
	}
	return nil
}
