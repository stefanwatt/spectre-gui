package neovimtransport

import (
	"nvim-gui/core/events"
	"nvim-gui/utils"
	"strconv"
	"strings"
)

// MapRedrawBatch translates raw nvim redraw updates into canonical events.
// Unknown or malformed entries are dropped.
func MapRedrawBatch(updates [][]interface{}) []events.Event {
	result := make([]events.Event, 0)
	for _, update := range updates {
		if len(update) == 0 {
			continue
		}
		eventName, ok := update[0].(string)
		if !ok {
			continue
		}
		args := update[1:]
		mapped := mapEvent(eventName, args)
		result = append(result, mapped...)
	}
	return result
}

func mapEvent(name string, args []interface{}) []events.Event {
	switch name {
	case string(events.EventFlush):
		return []events.Event{{Name: events.EventFlush, Source: "neovim-redraw"}}

	case string(events.EventGridResize):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 3 {
				continue
			}
			list = append(list, events.Event{
				Name: events.EventGridResize,
				Payload: events.GridResize{
					GridID: utils.ReflectToInt(a[0]),
					Width:  utils.ReflectToInt(a[1]),
					Height: utils.ReflectToInt(a[2]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventGridLine):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 4 {
				continue
			}
			cells, ok := a[3].([]interface{})
			if !ok {
				continue
			}
			list = append(list, events.Event{
				Name: events.EventGridLine,
				Payload: events.GridLine{
					GridID: utils.ReflectToInt(a[0]),
					Row:    utils.ReflectToInt(a[1]),
					Col:    utils.ReflectToInt(a[2]),
					Cells:  interfaceSlice(cells),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventGridClear):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 1 {
				continue
			}
			list = append(list, events.Event{
				Name:    events.EventGridClear,
				Payload: events.GridClear{GridID: utils.ReflectToInt(a[0])},
				Source:  "neovim-redraw",
			})
		}
		return list

	case string(events.EventGridScroll):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 7 {
				continue
			}
			list = append(list, events.Event{
				Name: events.EventGridScroll,
				Payload: events.GridScroll{
					GridID: utils.ReflectToInt(a[0]),
					Top:    utils.ReflectToInt(a[1]),
					Bottom: utils.ReflectToInt(a[2]),
					Left:   utils.ReflectToInt(a[3]),
					Right:  utils.ReflectToInt(a[4]),
					Rows:   utils.ReflectToInt(a[5]),
					Cols:   utils.ReflectToInt(a[6]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventGridCursorGoto):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 3 {
				continue
			}
			list = append(list, events.Event{
				Name: events.EventGridCursorGoto,
				Payload: events.GridCursorGoto{
					GridID: utils.ReflectToInt(a[0]),
					Row:    utils.ReflectToInt(a[1]),
					Col:    utils.ReflectToInt(a[2]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventModeChange):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 1 {
				continue
			}
			mode, _ := a[0].(string)
			list = append(list, events.Event{
				Name:    events.EventModeChange,
				Payload: events.ModeChange{Mode: mode},
				Source:  "neovim-redraw",
			})
		}
		return list

	case string(events.EventWinViewport):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 8 {
				continue
			}
			list = append(list, events.Event{
				Name: events.EventWinViewport,
				Payload: events.WindowViewport{
					GridID:      utils.ReflectToInt(a[0]),
					TopLine:     utils.ReflectToInt(a[2]),
					BottomLine:  utils.ReflectToInt(a[3]),
					CursorLine:  utils.ReflectToInt(a[4]),
					LineCount:   utils.ReflectToInt(a[6]),
					ScrollDelta: utils.ReflectToFloat(a[7]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventWinViewportMargin):
		if len(args) < 1 {
			return nil
		}
		a, ok := args[0].([]interface{})
		if !ok || len(a) < 6 {
			return nil
		}
		return []events.Event{{
			Name: events.EventWinViewportMargin,
			Payload: events.WindowViewportMargins{
				GridID: utils.ReflectToInt(a[0]),
				Top:    utils.ReflectToInt(a[2]),
				Bottom: utils.ReflectToInt(a[3]),
				Left:   utils.ReflectToInt(a[4]),
				Right:  utils.ReflectToInt(a[5]),
			},
			Source: "neovim-redraw",
		}}

	case string(events.EventWinPos):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 6 {
				continue
			}
			windowID, ok := mapWindowID(a[1])
			if !ok {
				continue
			}
			list = append(list, events.Event{
				Name: events.EventWinPos,
				Payload: events.WindowPosition{
					GridID:   utils.ReflectToInt(a[0]),
					WindowID: windowID,
					Row:      utils.ReflectToInt(a[2]),
					Col:      utils.ReflectToInt(a[3]),
					Width:    utils.ReflectToInt(a[4]),
					Height:   utils.ReflectToInt(a[5]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventWinFloatPos):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 8 {
				continue
			}
			windowID, ok := mapWindowID(a[1])
			if !ok {
				continue
			}
			anchor, _ := a[2].(string)
			focusable, _ := a[6].(bool)
			list = append(list, events.Event{
				Name: events.EventWinFloatPos,
				Payload: events.FloatingWindowPosition{
					GridID:     utils.ReflectToInt(a[0]),
					WindowID:   windowID,
					Anchor:     anchor,
					AnchorGrid: utils.ReflectToInt(a[3]),
					AnchorRow:  utils.ReflectToFloat(a[4]),
					AnchorCol:  utils.ReflectToFloat(a[5]),
					Focusable:  focusable,
					ZIndex:     utils.ReflectToInt(a[7]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventWinHide):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 1 {
				continue
			}
			list = append(list, events.Event{
				Name:    events.EventWinHide,
				Payload: utils.ReflectToInt(a[0]),
				Source:  "neovim-redraw",
			})
		}
		return list

	case string(events.EventWinClose):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 1 {
				continue
			}
			list = append(list, events.Event{
				Name:    events.EventWinClose,
				Payload: utils.ReflectToInt(a[0]),
				Source:  "neovim-redraw",
			})
		}
		return list

	case string(events.EventCmdlineShow):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 5 {
				continue
			}
			chunks, _ := a[0].([]interface{})
			firstc, _ := a[2].(string)
			prompt, _ := a[3].(string)
			list = append(list, events.Event{
				Name: events.EventCmdlineShow,
				Payload: events.CmdlineShow{
					Chunks: interfaceSlice(chunks),
					Pos:    utils.ReflectToInt(a[1]),
					Firstc: firstc,
					Prompt: prompt,
					Indent: utils.ReflectToInt(a[4]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventCmdlinePos):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 2 {
				continue
			}
			list = append(list, events.Event{
				Name: events.EventCmdlinePos,
				Payload: events.CmdlinePos{
					Pos:   utils.ReflectToInt(a[0]),
					Level: utils.ReflectToInt(a[1]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventCmdlineHide):
		return []events.Event{{
			Name:    events.EventCmdlineHide,
			Payload: events.CmdlineHide{},
			Source:  "neovim-redraw",
		}}

	case string(events.EventDefaultColorsSet):
		list := make([]events.Event, 0, len(args))
		for _, arg := range args {
			a, ok := arg.([]interface{})
			if !ok || len(a) < 3 {
				continue
			}
			list = append(list, events.Event{
				Name: events.EventDefaultColorsSet,
				Payload: events.DefaultColorsSet{
					FG: utils.ReflectToInt(a[0]),
					BG: utils.ReflectToInt(a[1]),
					SP: utils.ReflectToInt(a[2]),
				},
				Source: "neovim-redraw",
			})
		}
		return list

	case string(events.EventHighlightDefine):
		return []events.Event{{
			Name: events.EventHighlightDefine,
			Payload: events.HighlightAttrDefine{
				Args: interfaceSlice(args),
			},
			Source: "neovim-redraw",
		}}
	}

	return nil
}

func interfaceSlice(items []interface{}) []any {
	result := make([]any, len(items))
	for i := range items {
		result[i] = items[i]
	}
	return result
}

func mapWindowID(raw interface{}) (int, bool) {
	stringer, ok := raw.(interface{ String() string })
	if !ok {
		return 0, false
	}
	parts := strings.Split(stringer.String(), ":")
	if len(parts) != 2 {
		return 0, false
	}
	winID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}
	return winID, true
}
