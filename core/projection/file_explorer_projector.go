package projection

import (
	"nvim-gui/core/model"
	fileexplorer "nvim-gui/features/file-explorer"
	"nvim-gui/rendering"
)

type FileExplorerProjector struct {
	fileExplorer *fileexplorer.FileExplorer
	wasActive    bool // tracks previous state for close detection
}

func NewFileExplorerProjector(fileExplorer *fileexplorer.FileExplorer) *FileExplorerProjector {
	return &FileExplorerProjector{
		fileExplorer: fileExplorer,
		wasActive:    false,
	}
}

func (p *FileExplorerProjector) Project(state *model.AppState) UIProjection {
	s := &state.Editor.Screen
	active := p.fileExplorer.GetActive()
	// File explorer just closed — emit close event
	if !p.fileExplorer.GetActive() && p.wasActive {
		p.wasActive = false
		return UIProjection{Events: []EmittedEvent{
			{Name: "file-explorer-close", Payload: struct{}{}},
		}}
	}

	// Not active, nothing to do
	if !active {
		return UIProjection{Events: []EmittedEvent{}}
	}
	p.wasActive = true

	if !p.fileExplorer.Dirty {
		return UIProjection{Events: []EmittedEvent{}}
	}

	// Build the file-explorer-update payload
	events := []EmittedEvent{}

	payload := map[string]any{
		"parent":         p.fileExplorer.GetParent(),
		"current":        p.fileExplorer.GetCurrent(),
		"preview":        p.fileExplorer.GetPreview(),
		"currentWinMode": s.Mode,
	}

	events = append(events, EmittedEvent{Name: "file-explorer-update", Payload: payload})
	p.fileExplorer.Dirty = false

	return UIProjection{Events: events}
}

// buildDirectory parses a file explorer pane into a Directory struct.
// isCurrentPane indicates whether this is the focused "current" pane (which
// receives grid_cursor_goto) vs a non-focused pane like "parent".
// func (p *FileExplorerProjector) buildDirectory(
// 	s *model.ScreenState,
// 	winID, bufNr int,
// 	directoryLineIsDir map[int]map[int]bool,
// 	isCurrentPane bool,
// ) *fileexplorer.Directory {
// 	win, exists := s.Windows[winID]
// 	if !exists || !win.IsFileExplorer {
// 		return nil
// 	}
// 	grid := s.Grids[win.GridID]
// 	if grid == nil || grid.Height == 0 {
// 		return nil
// 	}
//
// 	lineMap := directoryLineIsDir[bufNr]
// 	if lineMap == nil {
// 		lineMap = map[int]bool{}
// 	}
//
// 	entries := fileexplorer.ParseDirectoryEntries(grid.Cells, grid.Height, lineMap)
//
// 	// Determine selected entry from cursor position.
// 	// The current pane is focused and receives grid_cursor_goto, so CursorRow
// 	// is reliable. Non-focused panes (parent) never get grid_cursor_goto, so
// 	// we use ViewportCursorLine from win_viewport which fires for all windows.
// 	// ViewportCursorLine is a 0-indexed buffer line; to convert to a grid row
// 	// we subtract TopLine (scroll offset) and add 1 for the top border row.
// 	selectedEntryId := -1
// 	cursorCol := 0
// 	cursorRow := grid.CursorRow
// 	if !isCurrentPane {
// 		cursorRow = grid.ViewportCursorLine - grid.TopLine + 1 // buffer line -> grid row (accounting for scroll + border)
// 	}
//
// 	// DEBUG: log cursor resolution for file explorer panes
// 	entryIDs := make([]uint64, len(entries))
// 	for i, e := range entries {
// 		entryIDs[i] = e.ID
// 	}
// 	log.Info(fmt.Sprintf("[fe-projector] buildDirectory winID=%d isCurrentPane=%v gridID=%d CursorRow=%d ViewportCursorLine=%d TopLine=%d cursorRow=%d entryIDs=%v",
// 		winID, isCurrentPane, win.GridID, grid.CursorRow, grid.ViewportCursorLine, grid.TopLine, cursorRow, entryIDs))
//
// 	if cursorRow >= 0 {
// 		for _, entry := range entries {
// 			if entry.ID == cursorRow {
// 				selectedEntryId = entry.ID
// 				break
// 			}
// 		}
// 		cursorCol = grid.CursorCol
// 	}
//
// 	return &fileexplorer.Directory{
// 		WinID:           winID,
// 		BufNr:           bufNr,
// 		Entries:         entries,
// 		SelectedEntryId: selectedEntryId,
// 		CursorCol:       cursorCol,
// 	}
// }

// buildPreview constructs the preview payload. If the currently selected entry
// is a directory, we parse it as a directory listing. Otherwise we render
// the grid as content rows (syntax-highlighted file preview).
func (p *FileExplorerProjector) buildPreview(
	s *model.ScreenState,
	snap fileexplorer.RegistrySnapshot,
	current *fileexplorer.Directory,
) any {
	win, exists := s.Windows[snap.Preview.WinID]
	if !exists || !win.IsFileExplorer {
		return nil
	}
	grid := s.Grids[win.GridID]
	if grid == nil || grid.Height == 0 {
		return nil
	}

	// Check if the currently selected entry in the "current" pane is a directory
	selectedIsDir := false
	if current != nil && current.SelectedEntryId >= 0 {
		for _, entry := range current.Entries {
			if entry.ID == current.SelectedEntryId {
				selectedIsDir = entry.IsDir
				break
			}
		}
	}

	if selectedIsDir {
		// Directory preview
		lineMap := snap.DirectoryLineIsDir[snap.Preview.BufNr]
		if lineMap == nil {
			lineMap = map[int]bool{}
		}
		entries := fileexplorer.ParseDirectoryEntries(grid.Cells, grid.Height, lineMap)
		return map[string]any{
			"directory": &fileexplorer.Directory{
				WinID:           snap.Preview.WinID,
				BufNr:           snap.Preview.BufNr,
				Entries:         entries,
				SelectedEntryId: 0,
				CursorCol:       0,
			},
		}
	}

	// Content preview — render via the standard content pipeline, then strip borders
	contentRows := p.renderContentPreview(grid)
	return map[string]any{
		"content": contentRows,
	}
}

// renderContentPreview builds ContentRow[] for a file preview grid,
// stripping the border (first/last row, first/last token per row).
func (p *FileExplorerProjector) renderContentPreview(grid *model.GridState) []rendering.ContentRow {
	// Copy cells to avoid aliasing the model's live slices.
	cellsCopy := make([][]*rendering.Cell, len(grid.Cells))
	for i, row := range grid.Cells {
		cellsCopy[i] = make([]*rendering.Cell, len(row))
		copy(cellsCopy[i], row)
	}
	gd := &rendering.GridData{
		Height:        grid.Height,
		TopLine:       grid.TopLine,
		DirtyRows:     make([]bool, grid.Height),
		Cells:         cellsCopy,
		OptimizedRows: make([][]*rendering.Cell, grid.Height),
		CachedTokens:  make([][]*rendering.Token, grid.Height),
		MarkdownOpts:  make(map[int]*rendering.MarkdownOpts),
		Cursor:        rendering.CursorPosition{Row: grid.CursorRow, Col: grid.CursorCol},
	}
	for i := 0; i < grid.Height; i++ {
		gd.DirtyRows[i] = true
	}

	output := rendering.BuildContentPayload(rendering.ContentInput{
		WindowID:   0,
		Filetype:   "minifiles",
		CursorLine: -1,
		Grid:       gd,
		Meta:       nil,
	})

	content := output.Content

	// Strip border: remove first and last row
	if len(content) > 2 {
		content = content[1 : len(content)-1]
	}
	// Strip border: remove first and last token from each row
	for i := range content {
		if len(content[i].Tokens) > 2 {
			content[i].Tokens = content[i].Tokens[1:]
			if len(content[i].Tokens) > 0 {
				content[i].Tokens = content[i].Tokens[:len(content[i].Tokens)-1]
			}
		}
	}

	return content
}
