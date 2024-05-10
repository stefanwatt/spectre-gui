package utils

import (
	"context"
	"path/filepath"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var watcher *fsnotify.Watcher

var debounced_files_changed = debounce.New(100 * time.Millisecond)

var start time.Time

func OnWrite(event fsnotify.Event, ctx context.Context) {
	path := event.Name
	if path[len(path)-1:] == "~" {
		return
	}
	Log("on_write")
	debounced_files_changed(func() {
		Runtime.EventsEmit(ctx, "files-changed")
	})
}

func OnDelete(event fsnotify.Event, ctx context.Context) {
	path := event.Name
	if path[len(path)-1:] == "~" {
		return
	}

	Log("on_delete")
	debounced_files_changed(func() {
		Runtime.EventsEmit(ctx, "files-changed")
	})
}

func ObserveFiles(ctx context.Context,
	dirs []string,
	dir string,
	on_write func(fsnotify.Event, context.Context),
	on_delete func(fsnotify.Event, context.Context),
) {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		Log("error creating filewatcher: " + err.Error())
		panic(err)
	}
	debouncedWrite := debounce.New(100 * time.Millisecond)
	debouncedDelete := debounce.New(100 * time.Millisecond)
	defer watcher.Close()
	err = watch_dirs(dirs)
	if err != nil {
		Log("error watching files: " + err.Error())
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

func watch_dirs(dirs []string) error {
	for _, dir := range dirs {
		Log("Adding watcher for " + dir)
		filepath.Dir(dir)
		err := watcher.Add(dir)
		if err != nil {
			return err
		}
	}
	return nil
}
