package filewatcher

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"spectre-gui/utils"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
)

var (
	watcher           *fsnotify.Watcher
	close_dir_watcher context.CancelFunc
)

func InitContext(current_dir string, new_dir string, ctx context.Context) (context.Context, string) {
	updated_dir := current_dir
	if current_dir != new_dir && close_dir_watcher != nil {
		close_dir_watcher()
		updated_dir = new_dir
	}
	ctx, cancel := context.WithCancel(ctx)
	close_dir_watcher = cancel
	return ctx, updated_dir
}

func WatchFiles(ctx context.Context,
	dirs []string,
	dir string,
	on_write func(fsnotify.Event, context.Context),
	on_delete func(fsnotify.Event, context.Context),
) error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		utils.Log("error creating filewatcher: " + err.Error())
		return err
	}
	defer watcher.Close()
	err = watch_dirs(dirs)
	if err != nil {
		utils.Log("error watching files: " + err.Error())
		return err
	}
	err = watcher.Add(dir)
	if err != nil {
		utils.Log("skipping directory " + dir + " because of error")
		utils.Log(err.Error())
		return err
	}
	debouncedWrite := debounce.New(100 * time.Millisecond)
	debouncedDelete := debounce.New(100 * time.Millisecond)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher.Events channel closed")
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
				return err
			}
			utils.Log("file watcher error")
			utils.Log(err.Error())
		}
	}
}

func watch_dirs(dirs []string) error {
	for _, dir := range dirs {
		filepath.Dir(dir)
		err := watcher.Add(dir)
		if err != nil {
			return err
		}
	}
	return nil
}
