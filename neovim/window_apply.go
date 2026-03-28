package neovim

import "github.com/charmbracelet/log"

func (s *Screen) ApplyWinPosMapped(p struct {
	GridID   int
	WindowID int
	Row      int
	Col      int
	Width    int
	Height   int
}) {
	grid, exists := s.Grids[p.GridID]
	if !exists {
		s.GridResize(p.GridID, p.Width, p.Height)
		grid = s.Grids[p.GridID]
	}
	if grid == nil {
		return
	}

	s.windowsMu.Lock()
	defer s.windowsMu.Unlock()

	window, exists := s.Windows[p.WindowID]
	if !exists {
		window = NewWindow(p.WindowID, grid)
		s.Windows[p.WindowID] = window
		EmitEvent("window_opened", struct{}{})
	}

	window.Grid = grid
	window.Type = "normal"
	window.StartRow = p.Row
	window.StartCol = p.Col
	window.Width = p.Width
	window.Height = p.Height
	window.Hidden = false
	window.Dirty = true
	window.lineNumbers = true
	window.relativeLineNumbers = true
	s.GridToWindow[p.GridID] = p.WindowID

	log.Debug("apply win_pos", "win_id", p.WindowID, "grid_id", p.GridID, "row", p.Row, "col", p.Col, "w", p.Width, "h", p.Height)
}

func (s *Screen) ApplyWinFloatPosMapped(p struct {
	GridID     int
	WindowID   int
	Anchor     string
	AnchorGrid int
	AnchorRow  float64
	AnchorCol  float64
	Focusable  bool
	ZIndex     int
}) {
	grid, exists := s.Grids[p.GridID]
	if !exists {
		s.GridResize(p.GridID, 1, 1)
		grid = s.Grids[p.GridID]
	}
	if grid == nil {
		return
	}

	s.windowsMu.Lock()
	defer s.windowsMu.Unlock()

	window, exists := s.Windows[p.WindowID]
	if !exists {
		window = NewWindow(p.WindowID, grid)
		s.Windows[p.WindowID] = window
	}

	window.Grid = grid
	window.Type = "floating"
	window.Anchor = p.Anchor
	window.AnchorGrid = p.AnchorGrid
	window.StartRow = int(p.AnchorRow)
	window.StartCol = int(p.AnchorCol)
	window.Focusable = p.Focusable
	window.ZIndex = p.ZIndex
	window.Hidden = false
	window.Dirty = true
	s.GridToWindow[p.GridID] = p.WindowID

	log.Debug("apply win_float_pos", "win_id", p.WindowID, "grid_id", p.GridID, "anchor", p.Anchor, "zindex", p.ZIndex)
}

func (s *Screen) HideWindowByID(winID int) {
	s.windowsMu.Lock()
	defer s.windowsMu.Unlock()
	window, exists := s.Windows[winID]
	if !exists {
		return
	}
	window.Hidden = true
	window.Dirty = false
	EmitEvent("hide-window", winID)
	EmitEvent("file-explorer-confirm-prompt-hide", struct{}{})
	log.Debug("apply win_hide", "win_id", winID)
}

func (s *Screen) CloseWindowByID(winID int) {
	s.windowsMu.Lock()
	defer s.windowsMu.Unlock()
	window, exists := s.Windows[winID]
	if !exists {
		return
	}
	if window.IsFloating() {
		if window.ZIndex == 69420 {
			EmitEvent("preview-window-closed", winID)
		} else if !window.IsFileExplorer {
			EmitEvent("floating_window_closed", winID)
		}
	}
	delete(s.GridToWindow, window.Grid.ID)
	delete(s.Windows, winID)
	EmitEvent("file-explorer-confirm-prompt-hide", struct{}{})
	log.Debug("apply win_close", "win_id", winID)
}
