package projection

import (
	"os"
	"path/filepath"
	"strings"

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

	parent := p.fileExplorer.GetParent()
	current := p.fileExplorer.GetCurrent()
	payload := map[string]any{
		"parent":         p.buildPanePayload(parent, p.parentTitle(parent)),
		"current":        p.buildPanePayload(current, p.basenameTitle(current.Path)),
		"preview":        p.buildPreviewPayload(state),
		"pathDisplay":    replaceHomePrefix(current.Path),
		"currentWinMode": s.Mode,
	}

	events = append(events, EmittedEvent{Name: "file-explorer-update", Payload: payload})
	p.fileExplorer.Dirty = false

	return UIProjection{Events: events}
}

func (p *FileExplorerProjector) buildPanePayload(directory fileexplorer.Directory, title string) map[string]any {
	return map[string]any{
		"winId":           directory.WinID,
		"bufNr":           directory.BufNr,
		"title":           title,
		"entries":         directory.Entries,
		"selectedEntryId": directory.SelectedEntryId,
		"cursorCol":       directory.CursorCol,
		"dirty":           p.fileExplorer.IsBufDirty(directory.BufNr),
	}
}

func (p *FileExplorerProjector) parentTitle(directory fileexplorer.Directory) string {
	return p.basenameTitle(directory.Path)
}

func (p *FileExplorerProjector) buildPreviewPayload(state *model.AppState) any {
	kind := p.fileExplorer.GetPreviewKind()
	if kind == fileexplorer.PreviewKindNone {
		return nil
	}
	title := p.fileExplorer.GetPreviewTitle()
	switch kind {
	case fileexplorer.PreviewKindDirectory:
		preview := p.fileExplorer.GetPreview()
		return map[string]any{
			"kind":      "directory",
			"title":     title,
			"directory": p.buildPanePayload(preview, title),
		}
	case fileexplorer.PreviewKindLocalImage:
		return map[string]any{
			"kind":  "localImage",
			"title": title,
			"path":  p.fileExplorer.GetPreviewPath(),
		}
	case fileexplorer.PreviewKindTextFile:
		content := []rendering.ContentRow{}
		winID := p.fileExplorer.GetPreviewWinID()
		if win, exists := state.Editor.Screen.Windows[winID]; exists {
			if grid := state.Editor.Screen.Grids[win.GridID]; grid != nil && grid.Height > 0 {
				content = p.renderContentPreview(grid, p.fileExplorer.GetPreviewFiletype())
			}
		}
		fallback := textLinesPreviewContent(p.fileExplorer.GetPreviewTextLines())
		if len(fallback) > 0 && (len(content) == 0 || !contentHasNonWhitespace(content)) {
			content = fallback
		}
		return map[string]any{
			"kind":    "textFile",
			"title":   title,
			"content": content,
		}
	default:
		return nil
	}
}

func textLinesPreviewContent(lines []string) []rendering.ContentRow {
	if len(lines) == 0 {
		return []rendering.ContentRow{}
	}
	content := make([]rendering.ContentRow, len(lines))
	for i, line := range lines {
		content[i] = rendering.ContentRow{
			Index: i + 1,
			Tokens: []*rendering.Token{{
				Text:      line,
				Classes:   "",
				Highlight: 0,
			}},
			Dirty: true,
		}
	}
	return content
}

func contentHasNonWhitespace(content []rendering.ContentRow) bool {
	for _, row := range content {
		for _, token := range row.Tokens {
			if strings.TrimSpace(token.Text) != "" {
				return true
			}
		}
	}
	return false
}

func (p *FileExplorerProjector) basenameTitle(path string) string {
	path = replaceHomePrefix(path)
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return filepath.Base(path)
}

func replaceHomePrefix(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}

	cleanPath := filepath.Clean(path)
	cleanHome := filepath.Clean(home)
	if cleanPath == cleanHome {
		return "~"
	}

	homePrefix := cleanHome + string(os.PathSeparator)
	if strings.HasPrefix(cleanPath, homePrefix) {
		return "~/" + strings.TrimPrefix(cleanPath, homePrefix)
	}
	return cleanPath
}

// renderContentPreview builds ContentRow[] for a file preview grid.
func (p *FileExplorerProjector) renderContentPreview(grid *model.GridState, filetype string) []rendering.ContentRow {
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
		Filetype:   filetype,
		CursorLine: -1,
		Grid:       gd,
		Meta:       nil,
	})

	return output.Content
}
