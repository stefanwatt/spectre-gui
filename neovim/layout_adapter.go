package neovim

import "nvim-gui/rendering"

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
		return
	}
	if rendering.LayoutEqual(s.layout, next) {
		return
	}
	s.layout = next
}
