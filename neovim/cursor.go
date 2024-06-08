package neovim

import (
	"context"
	"log"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type CursorMoveEvent struct {
	Row        uint64 `msgpack:"row" json:"row"`
	Col        uint64 `msgpack:"col" json:"col"`
	Key        string `msgpack:"key" json:"key"`
	TopLine    uint64 `msgpack:"top_line" json:"top_line"`
	BottomLine uint64 `msgpack:"bottom_line" json:"bottom_line"`
}

type NvimRange struct {
	StartRow uint64 `msgpack:"start_row" json:"start_row"`
	StartCol uint64 `msgpack:"start_col" json:"start_col"`
	EndRow   uint64 `msgpack:"end_row" json:"end_row"`
	EndCol   uint64 `msgpack:"end_col" json:"end_col"`
}

func UpdateCursor(ctx context.Context, cursor_move_event CursorMoveEvent) {
	log.Println("cursor moved ", cursor_move_event)
	if cursor_move_event.Key == "" {
		cursor_move_event.Key = " "
	}
	Runtime.EventsEmit(ctx, "cursor-changed", cursor_move_event)
}
