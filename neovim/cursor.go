package neovim

import (
	"fmt"
	"nvim-gui/utils"
)

type Cursor struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

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
			activeWin, exists := s.Windows[s.ActiveWindow]
			if exists {
				updatedCursor := Cursor{Row: windowCursor[0], Col: windowCursor[1]}
				activeWin.Cursor = &updatedCursor
			}
			cursorMoveEvent := CursorMoveEvent{
				Row:            uint64(windowCursor[0]),
				Col:            uint64(windowCursor[1]),
				ActiveWindowId: s.ActiveWindow,
			}
			s.emitEvent("cursor-changed", cursorMoveEvent)
		}
	} else {
		utils.Log(fmt.Sprintf("UpdateCursor could not get nvimWindow\nerror:%s", err.Error()))
	}

}
