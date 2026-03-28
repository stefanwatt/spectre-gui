package reducer

import (
	"nvim-gui/core/events"
	"nvim-gui/core/model"
)

func Apply(state *model.AppState, event events.Event) {
	s := &state.Editor.Screen

	switch event.Name {
	case events.EventModeChange:
		payload, ok := event.Payload.(events.ModeChange)
		if !ok {
			return
		}
		s.Mode = payload.Mode

	case events.EventGridResize:
		payload, ok := event.Payload.(events.GridResize)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			grid = &model.GridState{ID: payload.GridID, Lines: make(map[int]model.LineState)}
			s.Grids[payload.GridID] = grid
		}
		grid.Width = payload.Width
		grid.Height = payload.Height

	case events.EventGridCursorGoto:
		payload, ok := event.Payload.(events.GridCursorGoto)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			grid = &model.GridState{ID: payload.GridID, Lines: make(map[int]model.LineState)}
			s.Grids[payload.GridID] = grid
		}
		grid.CursorRow = payload.Row
		grid.CursorCol = payload.Col
		if winID, ok := s.GridToWindow[payload.GridID]; ok {
			s.ActiveWindow = winID
		}

	case events.EventGridLine:
		payload, ok := event.Payload.(events.GridLine)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			grid = &model.GridState{ID: payload.GridID, Lines: make(map[int]model.LineState)}
			s.Grids[payload.GridID] = grid
		}
		if grid.Lines == nil {
			grid.Lines = make(map[int]model.LineState)
		}
		grid.Lines[payload.Row] = model.LineState{
			Row:   payload.Row,
			Col:   payload.Col,
			Cells: payload.Cells,
		}

	case events.EventGridClear:
		payload, ok := event.Payload.(events.GridClear)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			return
		}
		grid.Lines = make(map[int]model.LineState)

	case events.EventGridScroll:
		payload, ok := event.Payload.(events.GridScroll)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			return
		}
		for row := payload.Top; row < payload.Bottom; row++ {
			line, exists := grid.Lines[row]
			if !exists {
				continue
			}
			newRow := row - payload.Rows
			if newRow < payload.Top || newRow >= payload.Bottom {
				delete(grid.Lines, row)
				continue
			}
			line.Row = newRow
			grid.Lines[newRow] = line
			if newRow != row {
				delete(grid.Lines, row)
			}
		}

	case events.EventWinPos:
		payload, ok := event.Payload.(events.WindowPosition)
		if !ok {
			return
		}
		win := s.Windows[payload.WindowID]
		if win == nil {
			win = &model.WindowState{ID: payload.WindowID}
			s.Windows[payload.WindowID] = win
		}
		win.Type = "normal"
		win.GridID = payload.GridID
		win.StartRow = payload.Row
		win.StartCol = payload.Col
		win.Width = payload.Width
		win.Height = payload.Height
		win.Hidden = false
		s.GridToWindow[payload.GridID] = payload.WindowID

	case events.EventWinFloatPos:
		payload, ok := event.Payload.(events.FloatingWindowPosition)
		if !ok {
			return
		}
		win := s.Windows[payload.WindowID]
		if win == nil {
			win = &model.WindowState{ID: payload.WindowID}
			s.Windows[payload.WindowID] = win
		}
		win.Type = "floating"
		win.GridID = payload.GridID
		win.Anchor = payload.Anchor
		win.AnchorGrid = payload.AnchorGrid
		win.StartRow = int(payload.AnchorRow)
		win.StartCol = int(payload.AnchorCol)
		win.Focusable = payload.Focusable
		win.ZIndex = payload.ZIndex
		s.GridToWindow[payload.GridID] = payload.WindowID

	case events.EventWinHide:
		payload, ok := event.Payload.(int)
		if !ok {
			return
		}
		if win, exists := s.Windows[payload]; exists {
			win.Hidden = true
		}

	case events.EventWinClose:
		payload, ok := event.Payload.(int)
		if !ok {
			return
		}
		if win, exists := s.Windows[payload]; exists {
			delete(s.GridToWindow, win.GridID)
			delete(s.Windows, payload)
		}

	case events.EventWinViewport:
		payload, ok := event.Payload.(events.WindowViewport)
		if !ok {
			return
		}
		s.Viewport.TopLine = payload.TopLine
		s.Viewport.BottomLine = payload.BottomLine
		s.Viewport.CursorLine = payload.CursorLine
		s.Viewport.LineCount = payload.LineCount
		s.Viewport.ScrollDelta = payload.ScrollDelta
		grid := s.Grids[payload.GridID]
		if grid == nil {
			grid = &model.GridState{ID: payload.GridID, Lines: make(map[int]model.LineState)}
			s.Grids[payload.GridID] = grid
		}
		grid.TopLine = payload.TopLine

	case events.EventWinViewportMargin:
		payload, ok := event.Payload.(events.WindowViewportMargins)
		if !ok {
			return
		}
		s.Viewport.Margins[0] = payload.Top
		s.Viewport.Margins[1] = payload.Bottom
		s.Viewport.Margins[2] = payload.Left
		s.Viewport.Margins[3] = payload.Right

	case events.EventCmdlineShow:
		payload, ok := event.Payload.(events.CmdlineShow)
		if !ok {
			return
		}
		s.Cmdline.Visible = true
		s.Cmdline.Pos = payload.Pos
		s.Cmdline.Firstc = payload.Firstc
		s.Cmdline.Prompt = payload.Prompt
		s.Cmdline.Indent = payload.Indent
		s.Cmdline.Chunks = payload.Chunks

	case events.EventCmdlinePos:
		payload, ok := event.Payload.(events.CmdlinePos)
		if !ok {
			return
		}
		s.Cmdline.Pos = payload.Pos
		s.Cmdline.Level = payload.Level

	case events.EventCmdlineHide:
		s.Cmdline.Visible = false

	case events.EventDefaultColorsSet:
		payload, ok := event.Payload.(events.DefaultColorsSet)
		if !ok {
			return
		}
		s.Highlights.DefaultFG = payload.FG
		s.Highlights.DefaultBG = payload.BG
		s.Highlights.DefaultSP = payload.SP

	case events.EventHighlightDefine:
		payload, ok := event.Payload.(events.HighlightAttrDefine)
		if !ok {
			return
		}
		s.Highlights.Definitions = append(s.Highlights.Definitions, payload.Args)
	}
}
