package picker

import (
	"context"
	"nvim-gui/utils"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var debounced_files_changed = debounce.New(100 * time.Millisecond)

var start time.Time

func OnWrite(event fsnotify.Event, ctx context.Context) {
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

func OnDelete(event fsnotify.Event, ctx context.Context) {
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
				rgLine: line,
				match:  nil,
			}
		})
		page := Page{
			index:   i,
			matches: matches,
		}
		pages = append(pages, page)
	}
	return Pagination{
		PageIndex: 0,
		pages:     pages,
	}
}
