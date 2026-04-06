package projection

import (
	"nvim-gui/core/model"
	"nvim-gui/rendering"
)

const previewWindowZIndex = 69420

// LayoutContentProjector computes layout and content projections.
// It maintains a cache of rendering.GridData per grid to support
// incremental (dirty-row-only) content updates.
type LayoutContentProjector struct {
	gridDataCache    map[int]*rendering.GridData
	lastLayout       *rendering.GridLayout
	lastFloatingWins map[int]int // winID -> zIndex, tracks previous floating windows for close detection
}

func NewLayoutContentProjector() *LayoutContentProjector {
	return &LayoutContentProjector{
		gridDataCache:    make(map[int]*rendering.GridData),
		lastFloatingWins: make(map[int]int),
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

	// --- Content for each visible normal window ---
	for _, win := range s.Windows {
		if win == nil || win.Hidden {
			continue
		}
		if win.Type == "floating" {
			continue
		}
		p.projectWindowContent(&ui, win, s)
	}

	// --- Floating windows ---
	p.projectFloatingWindows(&ui, s)

	// --- Cursor ---
	cursorRow := s.Viewport.CursorLine + 1
	cursorCol := 0
	if activeWin, exists := s.Windows[s.ActiveWindow]; exists {
		if grid, exists := s.Grids[activeWin.GridID]; exists {
			cursorCol = grid.CursorCol
			if cursorRow == 0 {
				cursorRow = grid.TopLine + grid.CursorRow + 1
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

// projectWindowContent builds and emits content-updated for a single window.
func (p *LayoutContentProjector) projectWindowContent(ui *UIProjection, win *model.WindowState, s *model.ScreenState) {
	grid := s.Grids[win.GridID]
	if grid == nil || grid.Height == 0 || grid.Width == 0 {
		return
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
		return
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

	// For non-preview floating windows, trim the border (perimeter)
	if win.Type == "floating" && win.ZIndex != previewWindowZIndex {
		output.Content = trimPerimeter(output.Content)
	}

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

// projectFloatingWindows handles all floating window events:
// - "floating_windows": list of non-preview floating windows
// - "preview-window": preview window (ZIndex == 69420)
// - "floating_window_closed" / "preview-window-closed": when floating windows disappear
// - "hide-window": when floating windows become hidden
// - "content-updated": grid content for floating windows
func (p *LayoutContentProjector) projectFloatingWindows(ui *UIProjection, s *model.ScreenState) {
	currentFloatingWins := make(map[int]int) // winID -> zIndex
	var floatingInfos []map[string]any
	hasDirtyFloating := false

	for winID, win := range s.Windows {
		if win == nil || win.Type != "floating" || win.IsFileExplorer {
			continue
		}

		// Track hidden floating windows — emit hide-window
		if win.Hidden {
			// If it was previously known and now hidden, emit hide-window
			if _, wasPrevious := p.lastFloatingWins[winID]; wasPrevious {
				ui.Events = append(ui.Events, EmittedEvent{
					Name:    "hide-window",
					Payload: winID,
				})
			}
			continue
		}

		currentFloatingWins[winID] = win.ZIndex

		// Skip blink-cmp-menu windows (handled by native completion)
		if win.Filetype == "blink-cmp-menu" {
			continue
		}

		grid := s.Grids[win.GridID]
		if grid == nil {
			continue
		}

		// Capture dirty state before projectWindowContent clears it
		isDirty := win.Dirty || gridHasDirtyRows(grid)

		// Emit content for dirty floating windows
		if isDirty {
			p.projectWindowContent(ui, win, s)
		}

		if win.ZIndex == previewWindowZIndex {
			// Preview windows are emitted separately
			if isDirty {
				previewInfo := mapFloatingWindowInfo(win, winID, s)
				ui.Events = append(ui.Events, EmittedEvent{
					Name:    "preview-window",
					Payload: previewInfo,
				})
			}
			continue
		}

		// Collect non-preview floating windows
		if isDirty {
			hasDirtyFloating = true
		}
		floatingInfos = append(floatingInfos, mapFloatingWindowInfo(win, winID, s))
	}

	// Emit full floating_windows list if any are dirty
	if hasDirtyFloating && len(floatingInfos) > 0 {
		ui.Events = append(ui.Events, EmittedEvent{
			Name:    "floating_windows",
			Payload: floatingInfos,
		})
	}

	// Detect closed floating windows by comparing with previous state
	for prevWinID, prevZIndex := range p.lastFloatingWins {
		if _, stillExists := currentFloatingWins[prevWinID]; !stillExists {
			if prevZIndex == previewWindowZIndex {
				ui.Events = append(ui.Events, EmittedEvent{
					Name:    "preview-window-closed",
					Payload: prevWinID,
				})
			} else {
				ui.Events = append(ui.Events, EmittedEvent{
					Name:    "floating_window_closed",
					Payload: prevWinID,
				})
			}
		}
	}

	p.lastFloatingWins = currentFloatingWins
}

// mapFloatingWindowInfo builds the payload map matching the frontend App.FloatingWindow interface.
func mapFloatingWindowInfo(win *model.WindowState, winID int, s *model.ScreenState) map[string]any {
	anchorWindow := 0
	if resolved, exists := s.GridToWindow[win.AnchorGrid]; exists {
		anchorWindow = resolved
	}
	return map[string]any{
		"id":           winID,
		"gridId":       win.GridID,
		"anchorWindow": anchorWindow,
		"anchor":       win.Anchor,
		"row":          win.StartRow,
		"col":          win.StartCol,
		"width":        win.Width,
		"height":       win.Height,
		"zIndex":       win.ZIndex,
		"focusable":    win.Focusable,
		"isPopup":      false, // TODO: derive from model when needed
		"isHex":        win.Filetype == "blink-cmp-menu",
		"filetype":     win.Filetype,
	}
}

// gridHasDirtyRows checks if any row in the grid is marked dirty.
func gridHasDirtyRows(grid *model.GridState) bool {
	for _, d := range grid.DirtyRows {
		if d {
			return true
		}
	}
	return false
}

// trimPerimeter removes the border characters from floating window content.
// Floating windows rendered by Neovim typically have a 1-cell border around the content.
func trimPerimeter(contentRows []rendering.ContentRow) []rendering.ContentRow {
	if len(contentRows) < 3 {
		return []rendering.ContentRow{}
	}

	// Remove first and last row (top and bottom border)
	contentRows = contentRows[1 : len(contentRows)-1]

	for i := range contentRows {
		if len(contentRows[i].Tokens) < 3 {
			contentRows[i].Tokens = []*rendering.Token{}
		} else {
			// Remove first and last token (left and right border)
			contentRows[i].Tokens = contentRows[i].Tokens[1 : len(contentRows[i].Tokens)-1]
		}
	}

	return contentRows
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

	// Sync cells — model.Cell and rendering.Cell are the same type (via alias).
	// Copy the inner slice to avoid aliasing (the model may mutate its rows
	// between projection cycles, e.g. during scroll or grid_line events).
	for row := 0; row < grid.Height; row++ {
		if row < len(grid.DirtyRows) && grid.DirtyRows[row] {
			gd.DirtyRows[row] = true
			if row < len(grid.Cells) {
				gd.Cells[row] = make([]*rendering.Cell, len(grid.Cells[row]))
				copy(gd.Cells[row], grid.Cells[row])
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
		isFloating := win.Type == "floating"
		result = append(result, rendering.LayoutWindow{
			ID:                  win.ID,
			Type:                win.Type,
			Hidden:              win.Hidden,
			Floating:            isFloating,
			StartRow:            win.StartRow,
			StartCol:            win.StartCol,
			Width:               win.Width,
			Height:              win.Height,
			LineNumbers:         !isFloating,
			RelativeLineNumbers: !isFloating,
			Mode:                s.Mode,
			Filetype:            win.Filetype,
			Filepath:            win.Filepath,
		})
	}
	return result
}
