package reducer

import (
	"nvim-gui/core/events"
	"nvim-gui/core/model"
	"nvim-gui/rendering"
	"strings"

	fileexplorer "nvim-gui/features/file-explorer"
)

type Reducer struct {
	fileExplorer *fileexplorer.FileExplorer
}

func NewReducer(fileExplorer *fileexplorer.FileExplorer) *Reducer {
	return &Reducer{
		fileExplorer,
	}
}

func (r *Reducer) Apply(state *model.AppState, event events.Event) {
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
			grid = &model.GridState{ID: payload.GridID}
			s.Grids[payload.GridID] = grid
		}
		grid.Width = payload.Width
		grid.Height = payload.Height
		grid.EnsureCells()
		if payload.GridID == 1 {
			s.Width = payload.Width
			s.Height = payload.Height
		}
		// Mark the window dirty if there is one
		markWindowDirtyByGrid(s, payload.GridID)

	case events.EventGridCursorGoto:
		payload, ok := event.Payload.(events.GridCursorGoto)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			grid = &model.GridState{ID: payload.GridID}
			grid.Width = 0
			grid.Height = 0
			s.Grids[payload.GridID] = grid
		}
		// Mark old cursor row dirty (to remove cursor highlight)
		grid.MarkRowDirty(grid.CursorRow)
		grid.CursorRow = payload.Row
		grid.CursorCol = payload.Col
		// Mark new cursor row dirty (to add cursor highlight)
		grid.MarkRowDirty(grid.CursorRow)
		if winID, ok := s.GridToWindow[payload.GridID]; ok {
			s.ActiveWindow = winID
		}
		markWindowDirtyByGrid(s, payload.GridID)

	case events.EventGridLine:
		payload, ok := event.Payload.(events.GridLine)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			grid = &model.GridState{ID: payload.GridID}
			s.Grids[payload.GridID] = grid
		}
		// Ensure cells are allocated
		if len(grid.Cells) == 0 && grid.Width > 0 && grid.Height > 0 {
			grid.EnsureCells()
		}
		// Parse raw cell arrays and populate the Cell grid
		applyGridLine(grid, payload.Row, payload.Col, payload.Cells)
		grid.MarkRowDirty(payload.Row)
		markWindowDirtyByGrid(s, payload.GridID)

	case events.EventGridClear:
		payload, ok := event.Payload.(events.GridClear)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			return
		}
		// Clear all cells to blank spaces
		for row := 0; row < grid.Height; row++ {
			for col := 0; col < grid.Width; col++ {
				if row < len(grid.Cells) && col < len(grid.Cells[row]) {
					grid.Cells[row][col] = &model.Cell{Char: " ", Highlight: 0, Classes: map[string]bool{}}
				}
			}
		}
		grid.MarkAllDirty()
		markWindowDirtyByGrid(s, payload.GridID)

	case events.EventGridScroll:
		payload, ok := event.Payload.(events.GridScroll)
		if !ok {
			return
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			return
		}
		applyGridScroll(grid, payload.Top, payload.Bottom, payload.Rows, payload.Left, payload.Right)
		markWindowDirtyByGrid(s, payload.GridID)

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
		win.Dirty = true
		s.GridToWindow[payload.GridID] = payload.WindowID
		if fp, ok := s.PendingFilepaths[payload.WindowID]; ok {
			win.Filepath = fp
			delete(s.PendingFilepaths, payload.WindowID)
		}

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
		win.Hidden = false
		win.Dirty = true
		s.GridToWindow[payload.GridID] = payload.WindowID
		// Set width/height from grid if available
		if grid := s.Grids[payload.GridID]; grid != nil {
			win.Width = grid.Width
			win.Height = grid.Height
		}

	case events.EventWinHide:
		payload, ok := event.Payload.(int)
		if !ok {
			return
		}
		if winID, exists := s.GridToWindow[payload]; exists {
			if win, ok := s.Windows[winID]; ok {
				win.Hidden = true
				win.Dirty = true
			}
		} else if win, exists := s.Windows[payload]; exists {
			win.Hidden = true
			win.Dirty = true
		}

	case events.EventWinClose:
		payload, ok := event.Payload.(int)
		if !ok {
			return
		}
		if winID, exists := s.GridToWindow[payload]; exists {
			if win, ok := s.Windows[winID]; ok {
				delete(s.GridToWindow, win.GridID)
				delete(s.Windows, winID)
			}
		} else if win, exists := s.Windows[payload]; exists {
			delete(s.GridToWindow, win.GridID)
			delete(s.Windows, payload)
		}
		hasAnyFileExplorer := false
		for _, w := range s.Windows {
			if w != nil && w.IsFileExplorer {
				hasAnyFileExplorer = true
				break
			}
		}
		state.Features.FileExplorer.Active = hasAnyFileExplorer

	case events.EventWinViewport:
		payload, ok := event.Payload.(events.WindowViewport)
		if !ok {
			return
		}
		// Only update global viewport from the active window's grid.
		// Floating windows (e.g. LSP progress) also emit win_viewport and
		// would overwrite the cursor line, causing relative line numbers in
		// the main window to flicker.
		if winID, ok := s.GridToWindow[payload.GridID]; ok && winID == s.ActiveWindow {
			s.Viewport.TopLine = payload.TopLine
			s.Viewport.BottomLine = payload.BottomLine
			s.Viewport.CursorLine = payload.CursorLine
			s.Viewport.LineCount = payload.LineCount
			s.Viewport.ScrollDelta = payload.ScrollDelta
		}
		grid := s.Grids[payload.GridID]
		if grid == nil {
			grid = &model.GridState{ID: payload.GridID}
			s.Grids[payload.GridID] = grid
		}
		grid.TopLine = payload.TopLine
		grid.ViewportCursorLine = payload.CursorLine
		markWindowDirtyByGrid(s, payload.GridID)

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
	case events.EventBufEnter:
		payload, ok := event.Payload.(events.BufEnter)
		if !ok {
			return
		}
		window := s.Windows[payload.WinID]
		if window != nil {
			window.Filepath = payload.Filepath
			window.Dirty = true
		} else {
			s.PendingFilepaths[payload.WinID] = payload.Filepath
		}

	case events.EventWindowBufferInfo:
		payload, ok := event.Payload.(events.WindowBufferInfo)
		if !ok {
			return
		}
		if win, exists := s.Windows[payload.WindowID]; exists {
			win.Filetype = payload.Filetype
			win.Filepath = payload.Filepath
			win.Dirty = true
		}

	case events.EventWindowOptions:
		payload, ok := event.Payload.(events.WindowOptions)
		if !ok {
			return
		}
		if win, exists := s.Windows[payload.WindowID]; exists {
			win.LineNumbers = payload.LineNumbers
			win.Dirty = true
		}

	}
}

// applyGridLine parses raw Neovim cell arrays and writes proper *Cell objects into the grid.
// Each cell in cells is []interface{}{text, [hlID], [repeat]}.
// The highlight ID is "sticky" — it carries forward from cell to cell within a grid_line event.
func applyGridLine(grid *model.GridState, row, startCol int, cells []any) {
	if row < 0 || row >= len(grid.Cells) {
		return
	}
	col := startCol
	lastHL := 0

	for _, raw := range cells {
		cell, ok := raw.([]interface{})
		if !ok || len(cell) == 0 {
			continue
		}
		text, _ := cell[0].(string)
		if len(cell) > 1 {
			lastHL = toInt(cell[1])
		}
		repeat := 1
		if len(cell) > 2 {
			repeat = toInt(cell[2])
		}
		if repeat < 1 {
			repeat = 1
		}

		for i := 0; i < repeat; i++ {
			if col >= len(grid.Cells[row]) {
				break
			}
			grid.Cells[row][col] = &model.Cell{
				Char:      text,
				Highlight: lastHL,
				Classes:   classesForHL(lastHL),
			}
			col++
		}
	}
}

// classesForHL resolves a highlight ID to a map[string]bool CSS class set.
func classesForHL(hlID int) map[string]bool {
	classStr := rendering.ClassesForHighlightID(hlID)
	classes := map[string]bool{}
	if classStr == "" {
		return classes
	}
	for _, c := range strings.Split(classStr, " ") {
		if c != "" {
			classes[c] = true
		}
	}
	return classes
}

// applyGridScroll shifts lines in the scrolling region.
// rows > 0 means scroll up (content moves up), rows < 0 means scroll down.
func applyGridScroll(grid *model.GridState, top, bottom, rows, left, right int) {
	if len(grid.Cells) == 0 {
		return
	}

	if rows > 0 {
		// Scroll up: iterate from top to bottom (forward)
		for row := top; row < bottom; row++ {
			srcRow := row + rows
			if srcRow >= top && srcRow < bottom && srcRow < len(grid.Cells) && row < len(grid.Cells) {
				if left == 0 && (right == 0 || right >= grid.Width) {
					// Full-width scroll: copy the row slice (not assign, to avoid aliasing)
					grid.Cells[row] = make([]*model.Cell, len(grid.Cells[srcRow]))
					copy(grid.Cells[row], grid.Cells[srcRow])
				} else {
					// Partial horizontal scroll
					for col := left; col < right && col < grid.Width; col++ {
						grid.Cells[row][col] = grid.Cells[srcRow][col]
					}
				}
			} else if row < len(grid.Cells) {
				// Source row is out of bounds, fill with blanks
				if left == 0 && (right == 0 || right >= grid.Width) {
					for col := 0; col < grid.Width; col++ {
						grid.Cells[row][col] = &model.Cell{Char: " ", Highlight: 0, Classes: map[string]bool{}}
					}
				} else {
					for col := left; col < right && col < grid.Width; col++ {
						grid.Cells[row][col] = &model.Cell{Char: " ", Highlight: 0, Classes: map[string]bool{}}
					}
				}
			}
			grid.MarkRowDirty(row)
		}
	} else if rows < 0 {
		// Scroll down: iterate from bottom to top (backward)
		for row := bottom - 1; row >= top; row-- {
			srcRow := row + rows // rows is negative, so srcRow < row
			if srcRow >= top && srcRow < bottom && srcRow < len(grid.Cells) && row < len(grid.Cells) {
				if left == 0 && (right == 0 || right >= grid.Width) {
					// Full-width scroll: copy the row slice (not assign, to avoid aliasing)
					grid.Cells[row] = make([]*model.Cell, len(grid.Cells[srcRow]))
					copy(grid.Cells[row], grid.Cells[srcRow])
				} else {
					for col := left; col < right && col < grid.Width; col++ {
						grid.Cells[row][col] = grid.Cells[srcRow][col]
					}
				}
			} else if row < len(grid.Cells) {
				if left == 0 && (right == 0 || right >= grid.Width) {
					for col := 0; col < grid.Width; col++ {
						grid.Cells[row][col] = &model.Cell{Char: " ", Highlight: 0, Classes: map[string]bool{}}
					}
				} else {
					for col := left; col < right && col < grid.Width; col++ {
						grid.Cells[row][col] = &model.Cell{Char: " ", Highlight: 0, Classes: map[string]bool{}}
					}
				}
			}
			grid.MarkRowDirty(row)
		}
	}
}

// markWindowDirtyByGrid marks the window associated with a grid as dirty.
func markWindowDirtyByGrid(s *model.ScreenState, gridID int) {
	if winID, exists := s.GridToWindow[gridID]; exists {
		if win, ok := s.Windows[winID]; ok {
			win.Dirty = true
		}
	}
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case uint64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}
