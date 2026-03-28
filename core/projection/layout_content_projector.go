package projection

import (
	"nvim-gui/core/model"
	"nvim-gui/rendering"
)

// LayoutContentProjector computes layout and content projections.
// It maintains a cache of rendering.GridData per grid to support
// incremental (dirty-row-only) content updates.
type LayoutContentProjector struct {
	gridDataCache map[int]*rendering.GridData
	lastLayout    *rendering.GridLayout
}

func NewLayoutContentProjector() *LayoutContentProjector {
	return &LayoutContentProjector{
		gridDataCache: make(map[int]*rendering.GridData),
	}
}

func (p *LayoutContentProjector) Project(state *model.AppState) UIProjection {
	s := &state.Editor.Screen
	ui := UIProjection{Events: []EmittedEvent{}}

	// --- Layout ---
	layout := rendering.CalculateGridLayout(rendering.LayoutInput{
		ScreenWidth:    s.Width,
		ScreenHeight:   s.Height,
		ActiveWindowID: s.ActiveWindow,
		Windows:        mapLayoutWindows(s),
	})

	// Only emit layout-updated if the layout actually changed
	if p.lastLayout == nil || !rendering.LayoutEqual(p.lastLayout, layout) {
		ui.Events = append(ui.Events, EmittedEvent{Name: "layout-updated", Payload: layout})
		p.lastLayout = layout
	}

	// --- Content for each visible window ---
	for _, win := range s.Windows {
		if win == nil || win.Hidden {
			continue
		}
		if win.Type == "floating" {
			continue
		}
		grid := s.Grids[win.GridID]
		if grid == nil || grid.Height == 0 || grid.Width == 0 {
			continue
		}

		// Check if any rows are dirty for this grid
		hasDirty := false
		for _, d := range grid.DirtyRows {
			if d {
				hasDirty = true
				break
			}
		}
		if !hasDirty && !win.Dirty {
			continue
		}

		// Get or create cached GridData for this grid
		gd := p.getOrCreateGridData(win.GridID, grid)

		// Sync dirty rows from model to GridData
		syncGridData(gd, grid)

		// Build content payload using the existing pure rendering pipeline
		output := rendering.BuildContentPayload(rendering.ContentInput{
			WindowID:   win.ID,
			Filetype:   win.Filetype,
			CursorLine: -1, // no markdown cursor suppression for now
			Grid:       gd,
			Meta:       nil, // no markdown meta for now
		})

		ui.Events = append(ui.Events, EmittedEvent{
			Name:    "content-updated",
			Payload: output,
		})

		// Reset dirty flags on the model
		for i := range grid.DirtyRows {
			grid.DirtyRows[i] = false
		}
		win.Dirty = false
	}

	// --- Cursor ---
	// Emit cursor position using the active window's grid cursor data.
	// The frontend expects buffer-line coordinates (1-indexed row from WindowCursor RPC).
	// We approximate using viewport.CursorLine when available, falling back to grid cursor + topLine.
	cursorRow := s.Viewport.CursorLine
	cursorCol := 0
	if activeWin, exists := s.Windows[s.ActiveWindow]; exists {
		if grid, exists := s.Grids[activeWin.GridID]; exists {
			cursorCol = grid.CursorCol
			if cursorRow == 0 {
				cursorRow = grid.TopLine + grid.CursorRow
			}
		}
	}
	ui.Events = append(ui.Events, EmittedEvent{
		Name: "cursor-changed",
		Payload: map[string]any{
			"row":            uint64(cursorRow),
			"col":            uint64(cursorCol),
			"activeWindowId": s.ActiveWindow,
		},
	})

	// --- Mode ---
	ui.Events = append(ui.Events, EmittedEvent{
		Name:    "mode-changed",
		Payload: s.Mode,
	})

	return ui
}

// getOrCreateGridData returns a cached GridData or creates a new one.
func (p *LayoutContentProjector) getOrCreateGridData(gridID int, grid *model.GridState) *rendering.GridData {
	gd, exists := p.gridDataCache[gridID]
	if !exists || gd.Height != grid.Height {
		gd = &rendering.GridData{
			Height:        grid.Height,
			TopLine:       grid.TopLine,
			DirtyRows:     make([]bool, grid.Height),
			Cells:         make([][]*rendering.Cell, grid.Height),
			OptimizedRows: make([][]*rendering.Cell, grid.Height),
			CachedTokens:  make([][]*rendering.Token, grid.Height),
			MarkdownOpts:  make(map[int]*rendering.MarkdownOpts),
		}
		for row := 0; row < grid.Height; row++ {
			gd.DirtyRows[row] = true
		}
		p.gridDataCache[gridID] = gd
	}
	return gd
}

// syncGridData copies model grid state into the rendering GridData.
func syncGridData(gd *rendering.GridData, grid *model.GridState) {
	gd.TopLine = grid.TopLine
	gd.Height = grid.Height
	gd.Cursor = rendering.CursorPosition{Row: grid.CursorRow, Col: grid.CursorCol}

	// Sync cells — model.Cell and rendering.Cell are the same type (via alias)
	for row := 0; row < grid.Height; row++ {
		if row < len(grid.DirtyRows) && grid.DirtyRows[row] {
			gd.DirtyRows[row] = true
			if row < len(grid.Cells) {
				gd.Cells[row] = grid.Cells[row]
			}
		}
	}
}

func mapLayoutWindows(s *model.ScreenState) []rendering.LayoutWindow {
	result := make([]rendering.LayoutWindow, 0, len(s.Windows))
	for _, win := range s.Windows {
		if win == nil {
			continue
		}
		result = append(result, rendering.LayoutWindow{
			ID:       win.ID,
			Type:     win.Type,
			Hidden:   win.Hidden,
			Floating: win.Type == "floating",
			StartRow: win.StartRow,
			StartCol: win.StartCol,
			Width:    win.Width,
			Height:   win.Height,
			Mode:     s.Mode,
			Filetype: win.Filetype,
			Filepath: win.Filepath,
		})
	}
	return result
}
