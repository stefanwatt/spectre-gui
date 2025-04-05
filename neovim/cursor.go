package neovim

import (
	"fmt"
	"nvim-gui/utils"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type CursorMoveEvent struct {
	Row            uint64 `msgpack:"row" json:"row"`
	Col            uint64 `msgpack:"col" json:"col"`
	TopLine        uint64 `msgpack:"top_line" json:"top_line"`
	BottomLine     uint64 `msgpack:"bottom_line" json:"bottom_line"`
	ActiveWindowId int    `json:"activeWindowId"`
}

type NvimRange struct {
	StartRow uint64 `msgpack:"start_row" json:"start_row"`
	StartCol uint64 `msgpack:"start_col" json:"start_col"`
	EndRow   uint64 `msgpack:"end_row" json:"end_row"`
	EndCol   uint64 `msgpack:"end_col" json:"end_col"`
}

func (s *Screen) UpdateCursor() {
	nvimWindow, err := getWindow(s.ActiveWindow)
	if err == nil && nvimWindow != nil {
		windowCursor, err := NvimInstance.WindowCursor(*nvimWindow)
		if err == nil {
			cursorMoveEvent := CursorMoveEvent{
				Row:            uint64(windowCursor[0]),
				Col:            uint64(windowCursor[1]),
				ActiveWindowId: s.ActiveWindow,
			}
			Runtime.EventsEmit(s.ctx, "cursor-changed", cursorMoveEvent)
		}
	} else {
		utils.Log(fmt.Sprintf("UpdateCursor could not get nvimWindow\nerror:%s",err.Error()))
	}

}
