package neovim

import (
	"nvim-gui/rendering"

	"github.com/charmbracelet/log"
)

func (s *Screen) buildLayoutInput() rendering.LayoutInput {
	windows := make([]rendering.LayoutWindow, 0, len(s.Windows))
	for winID, window := range s.Windows {
		filetype := ""
		filepath := ""
		if window.Buffer != nil {
			filetype = window.Buffer.Filetype
			filepath = window.Buffer.Filepath
		}
		windows = append(windows, rendering.LayoutWindow{
			ID:                  winID,
			Type:                window.Type,
			Hidden:              window.Hidden,
			Floating:            window.IsFloating(),
			StartRow:            window.StartRow,
			StartCol:            window.StartCol,
			Width:               window.Width,
			Height:              window.Height,
			LineNumbers:         window.lineNumbers,
			RelativeLineNumbers: window.relativeLineNumbers,
			Mode:                window.Mode,
			Cursor:              mapCursor(window.Cursor),
			Filetype:            filetype,
			Filepath:            filepath,
		})
	}

	return rendering.LayoutInput{
		ScreenWidth:       s.Width,
		ScreenHeight:      s.Height,
		ActiveWindowID:    s.ActiveWindow,
		ColorColumns:      s.ColorColumns,
		ColorColumnColor:  s.ColorColumnColor,
		CursorLineEnabled: s.CursorLineEnabled,
		CursorLineColor:   s.CursorLineColor,
		Windows:           windows,
	}
}

func mapCursor(c *Cursor) *rendering.Cursor {
	if c == nil {
		return nil
	}
	return &rendering.Cursor{Row: c.Row, Col: c.Col}
}

func (s *Screen) CalculateGridLayout() {
	next := rendering.CalculateGridLayout(s.buildLayoutInput())
	if s.layout == nil {
		s.layout = next
		log.Debug("layout calculated (init)", "windows", len(next.Windows), "cols", next.Cols, "rows", next.Rows, "active_window", next.ActiveWindowId)
		return
	}
	if rendering.LayoutEqual(s.layout, next) {
		log.Debug("layout unchanged", "windows", len(next.Windows), "active_window", next.ActiveWindowId)
		return
	}
	log.Debug("layout changed", "old_windows", len(s.layout.Windows), "new_windows", len(next.Windows), "old_active", s.layout.ActiveWindowId, "new_active", next.ActiveWindowId)
	s.layout = next
}
