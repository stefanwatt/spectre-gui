package neovim

import (
	"nvim-gui/rendering"
	"nvim-gui/utils"
)

var (
	screen   *Screen
	renderer *rendering.Renderer
)

func HandleRedraw(updates [][]interface{}) {
	for _, update := range updates {
		if len(update) == 0 {
			continue
		}
		event, ok := handleEvent(update[0])
		if !ok {
			continue
		}

		args := update[1:]

		switch event {
		case "flush":
			renderer.ScheduleRender()
		case "grid_resize":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(gridArgs[0])
				width := utils.ReflectToInt(gridArgs[1])
				height := utils.ReflectToInt(gridArgs[2])
				screen.GridResize(gridId, width, height)
			}
		case "grid_line":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(gridArgs[0])
				row := utils.ReflectToInt(gridArgs[1])
				col := utils.ReflectToInt(gridArgs[2])
				cells := gridArgs[3].([]interface{})
				screen.GridLine(gridId, row, col, cells)
			}
		case "grid_clear":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(gridArgs[0])
				screen.GridClear(gridId)
			}
		case "grid_scroll":
			for _, arg := range args {
				scrollArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(scrollArgs[0])
				top := utils.ReflectToInt(scrollArgs[1])
				bot := utils.ReflectToInt(scrollArgs[2])
				left := utils.ReflectToInt(scrollArgs[3])
				right := utils.ReflectToInt(scrollArgs[4])
				rows := utils.ReflectToInt(scrollArgs[5])
				cols := utils.ReflectToInt(scrollArgs[6])
				screen.GridScroll(gridId, top, bot, left, right, rows, cols)
			}
		case "grid_cursor_goto":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(gridArgs[0])
				row := utils.ReflectToInt(gridArgs[1])
				col := utils.ReflectToInt(gridArgs[2])
				screen.GridCursorGoto(gridId, row, col)
			}
		case "win_viewport":
			screen.HandleWinViewport(args)
		case "win_viewport_margins":
			screen.HandleWinViewportMargins(args)
		case "default_colors_set":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				fg := utils.ReflectToInt(gridArgs[0])
				bg := utils.ReflectToInt(gridArgs[1])
				sp := utils.ReflectToInt(gridArgs[2])
				DefaultColorsSet(fg, bg, sp)
				screen.MarkAllWindowsDirty()
				renderer.ScheduleRender()
			}
		case "hl_attr_define":
			HlAttrDefine(args)
			css := rendering.BuildHighlightCSS()
			EmitEvent("highlight-css", css)
			screen.MarkAllWindowsDirty()
			renderer.ScheduleRender()
		case "mode_change":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				mode, _ := gridArgs[0].(string)
				screen.ModeChange(mode)
			}
		case "win_pos":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				screen.WinPos(gridArgs)
			}
		case "win_float_pos":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				screen.WinFloatPos(gridArgs)
			}
		case "win_close":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				screen.WinClose(gridArgs)
			}

		case "win_hide":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				screen.WinHide(gridArgs)
			}
		case "cmdline_show":
			screen.HandleCmdlineShow(args)
		case "cmdline_pos":
			screen.HandleCmdlinePos(args)
		case "cmdline_hide":
			EmitEvent("cmdline_hide", struct{}{})
		}
	}
}

func handleEvent(update interface{}) (event string, ok bool) {
	switch update.(type) {
	case string:
		event = update.(string)
		ok = true
	default:
		event = ""
		ok = false
	}

	return event, ok
}
