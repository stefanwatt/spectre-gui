package projection

import (
	"nvim-gui/core/model"
	fileexplorer "nvim-gui/features/file-explorer"
	"nvim-gui/rendering"
)

type FileExplorerProjector struct {
	fileExplorerRegistry *fileexplorer.Registry
	wasActive            bool // tracks previous state for close detection
	wasConfirmVisible    bool // tracks confirm prompt visibility
}

func NewFileExplorerProjector(fileExplorerRegistry *fileexplorer.Registry) *FileExplorerProjector {
	return &FileExplorerProjector{
		fileExplorerRegistry: fileExplorerRegistry,
		wasActive:            false,
	}
}

func (p *FileExplorerProjector) Project(state *model.AppState) UIProjection {
	snap := p.fileExplorerRegistry.Snapshot()
	s := &state.Editor.Screen

	// File explorer just closed — emit close event
	if !snap.Active && p.wasActive {
		p.wasActive = false
		p.wasConfirmVisible = false
		return UIProjection{Events: []EmittedEvent{
			{Name: "file-explorer-close", Payload: struct{}{}},
			{Name: "file-explorer-confirm-prompt-hide", Payload: struct{}{}},
		}}
	}

	// Not active, nothing to do
	if !snap.Active {
		return UIProjection{Events: []EmittedEvent{}}
	}
	p.wasActive = true

	// Check if any file explorer window is dirty
	hasDirty := false
	for _, win := range s.Windows {
		if win != nil && win.IsFileExplorer && win.Dirty {
			hasDirty = true
			break
		}
	}
	if !hasDirty {
		// Still check non-FE floating windows for confirm prompt changes
		return p.projectConfirmPrompt(s)
	}

	// Build the file-explorer-update payload
	events := []EmittedEvent{}

	// Parse parent pane
	var parent *fileexplorer.Directory
	if snap.Parent.WinID > 0 {
		parent = p.buildDirectory(s, snap.Parent.WinID, snap.Parent.BufNr, snap.DirectoryLineIsDir)
	}

	// Parse current pane
	var current *fileexplorer.Directory
	if snap.Current.WinID > 0 {
		current = p.buildDirectory(s, snap.Current.WinID, snap.Current.BufNr, snap.DirectoryLineIsDir)
	}

	// Parse preview pane — can be directory or content
	var preview any
	if snap.Preview.WinID > 0 {
		preview = p.buildPreview(s, snap, current)
	}

	payload := map[string]any{
		"parent":         parent,
		"current":        current,
		"preview":        preview,
		"currentWinMode": s.Mode,
	}

	events = append(events, EmittedEvent{Name: "file-explorer-update", Payload: payload})

	// Also check for confirm prompt
	confirmEvents := p.projectConfirmPrompt(s)
	events = append(events, confirmEvents.Events...)

	return UIProjection{Events: events}
}

// buildDirectory parses a file explorer pane into a Directory struct.
func (p *FileExplorerProjector) buildDirectory(
	s *model.ScreenState,
	winID, bufNr int,
	directoryLineIsDir map[int]map[int]bool,
) *fileexplorer.Directory {
	win, exists := s.Windows[winID]
	if !exists || !win.IsFileExplorer {
		return nil
	}
	grid := s.Grids[win.GridID]
	if grid == nil || grid.Height == 0 {
		return nil
	}

	lineMap := directoryLineIsDir[bufNr]
	if lineMap == nil {
		lineMap = map[int]bool{}
	}

	entries := fileexplorer.ParseDirectoryEntries(grid.Cells, grid.Height, lineMap)

	// Determine selected entry from cursor position
	selectedEntryId := -1
	cursorCol := 0
	if grid.CursorRow >= 0 {
		for _, entry := range entries {
			if entry.ID == grid.CursorRow {
				selectedEntryId = entry.ID
				break
			}
		}
		cursorCol = grid.CursorCol
	}

	return &fileexplorer.Directory{
		WinID:           winID,
		BufNr:           bufNr,
		Entries:         entries,
		SelectedEntryId: selectedEntryId,
		CursorCol:       cursorCol,
	}
}

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
				SelectedEntryId: -1,
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
	gd := &rendering.GridData{
		Height:        grid.Height,
		TopLine:       grid.TopLine,
		DirtyRows:     make([]bool, grid.Height),
		Cells:         grid.Cells,
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

// projectConfirmPrompt scans non-file-explorer floating windows for the
// mini.files "close without synchronization" confirm dialog.
func (p *FileExplorerProjector) projectConfirmPrompt(s *model.ScreenState) UIProjection {
	events := []EmittedEvent{}

	found := false
	for winID, win := range s.Windows {
		if win == nil || win.Type != "floating" || win.IsFileExplorer || win.Hidden {
			continue
		}
		grid := s.Grids[win.GridID]
		if grid == nil || grid.Height == 0 {
			continue
		}
		prompt, ok := fileexplorer.DetectConfirmPrompt(winID, grid.Cells, win.Filetype)
		if ok {
			found = true
			if !p.wasConfirmVisible {
				p.wasConfirmVisible = true
				events = append(events, EmittedEvent{
					Name:    "file-explorer-confirm-prompt-show",
					Payload: prompt,
				})
			}
			break
		}
	}

	if !found && p.wasConfirmVisible {
		p.wasConfirmVisible = false
		events = append(events, EmittedEvent{
			Name:    "file-explorer-confirm-prompt-hide",
			Payload: struct{}{},
		})
	}

	return UIProjection{Events: events}
}
