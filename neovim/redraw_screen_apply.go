package neovim

import "nvim-gui/core/events"

func applyMappedEventToScreen(ev events.Event) {
	if NvimScreen == nil {
		return
	}

	switch ev.Name {
	case events.EventGridResize:
		if p, ok := ev.Payload.(events.GridResize); ok {
			NvimScreen.GridResize(p.GridID, p.Width, p.Height)
		}
	case events.EventGridLine:
		if p, ok := ev.Payload.(events.GridLine); ok {
			cells := make([]interface{}, len(p.Cells))
			for i := range p.Cells {
				cells[i] = p.Cells[i]
			}
			NvimScreen.GridLine(p.GridID, p.Row, p.Col, cells)
		}
	case events.EventGridClear:
		if p, ok := ev.Payload.(events.GridClear); ok {
			NvimScreen.GridClear(p.GridID)
		}
	case events.EventGridScroll:
		if p, ok := ev.Payload.(events.GridScroll); ok {
			NvimScreen.GridScroll(p.GridID, p.Top, p.Bottom, p.Left, p.Right, p.Rows, p.Cols)
		}
	case events.EventGridCursorGoto:
		if p, ok := ev.Payload.(events.GridCursorGoto); ok {
			NvimScreen.GridCursorGoto(p.GridID, p.Row, p.Col)
		}
	case events.EventModeChange:
		if p, ok := ev.Payload.(events.ModeChange); ok {
			NvimScreen.ModeChange(p.Mode)
		}
	case events.EventWinViewport:
		if p, ok := ev.Payload.(events.WindowViewport); ok {
			NvimScreen.HandleWinViewport([]interface{}{[]interface{}{p.GridID, 0, p.TopLine, p.BottomLine, p.CursorLine, 0, p.LineCount, p.ScrollDelta}})
		}
	case events.EventWinViewportMargin:
		if p, ok := ev.Payload.(events.WindowViewportMargins); ok {
			NvimScreen.HandleWinViewportMargins([]interface{}{[]interface{}{p.GridID, 0, p.Top, p.Bottom, p.Left, p.Right}})
		}
	case events.EventWinPos:
		if p, ok := ev.Payload.(events.WindowPosition); ok {
			NvimScreen.ApplyWinPosMapped(struct {
				GridID   int
				WindowID int
				Row      int
				Col      int
				Width    int
				Height   int
			}{
				GridID:   p.GridID,
				WindowID: p.WindowID,
				Row:      p.Row,
				Col:      p.Col,
				Width:    p.Width,
				Height:   p.Height,
			})
		}
	case events.EventWinFloatPos:
		if p, ok := ev.Payload.(events.FloatingWindowPosition); ok {
			NvimScreen.ApplyWinFloatPosMapped(struct {
				GridID     int
				WindowID   int
				Anchor     string
				AnchorGrid int
				AnchorRow  float64
				AnchorCol  float64
				Focusable  bool
				ZIndex     int
			}{
				GridID:     p.GridID,
				WindowID:   p.WindowID,
				Anchor:     p.Anchor,
				AnchorGrid: p.AnchorGrid,
				AnchorRow:  p.AnchorRow,
				AnchorCol:  p.AnchorCol,
				Focusable:  p.Focusable,
				ZIndex:     p.ZIndex,
			})
		}
	case events.EventWinHide:
		if winID, ok := ev.Payload.(int); ok {
			NvimScreen.HideWindowByID(winID)
		}
	case events.EventWinClose:
		if winID, ok := ev.Payload.(int); ok {
			NvimScreen.CloseWindowByID(winID)
		}
	}
}
