package events

type Name string

const (
	EventFlush             Name = "flush"
	EventGridResize        Name = "grid_resize"
	EventGridLine          Name = "grid_line"
	EventGridClear         Name = "grid_clear"
	EventGridScroll        Name = "grid_scroll"
	EventGridCursorGoto    Name = "grid_cursor_goto"
	EventWinPos            Name = "win_pos"
	EventWinFloatPos       Name = "win_float_pos"
	EventWinClose          Name = "win_close"
	EventWinHide           Name = "win_hide"
	EventWinViewport       Name = "win_viewport"
	EventWinViewportMargin Name = "win_viewport_margins"
	EventModeChange        Name = "mode_change"
	EventCmdlineShow       Name = "cmdline_show"
	EventCmdlinePos        Name = "cmdline_pos"
	EventCmdlineHide       Name = "cmdline_hide"
	EventHighlightDefine   Name = "hl_attr_define"
	EventDefaultColorsSet  Name = "default_colors_set"
	EventBufEnter          Name = "BufEnter"
	EventWindowBufferInfo  Name = "window_buffer_info"
)

type Event struct {
	Name    Name
	Payload any
	Source  string
}
