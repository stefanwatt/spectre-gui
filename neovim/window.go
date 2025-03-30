package neovim

// Window represents a Neovim window, which can be a regular window or a floating window
type Window struct {
	ID          int    // Window ID
	GridID      int    // Associated grid ID
	Type        string // "normal" or "floating"
	Anchor      string // Anchor position for floating windows
	AnchorGrid  int    // Grid ID this window is anchored to
	Row         int    // Row position
	Col         int    // Column position
	Width       int    // Window width
	Height      int    // Window height
	Focusable   bool   // Whether the window can be focused
	ZIndex      int    // Z-index for floating windows
	IsPopupmenu bool   // Whether this is a completion menu
}

// NewWindow creates a new window with the given ID and grid ID
func NewWindow(id int, gridID int) *Window {
	return &Window{
		ID:          id,
		GridID:      gridID,
		Type:        "normal",
		Anchor:      "",
		AnchorGrid:  0,
		Row:         0,
		Col:         0,
		Width:       0,
		Height:      0,
		Focusable:   true,
		ZIndex:      0,
		IsPopupmenu: false,
	}
}

// IsFloating returns true if this is a floating window
func (w *Window) IsFloating() bool {
	return w.Type == "floating"
}

// IsCompletionWindow tries to determine if this window is a completion menu
// based on its characteristics
func (w *Window) IsCompletionWindow() bool {
	if w.IsPopupmenu {
		return true
	}

	// Heuristics to identify completion windows:
	// - Must be floating
	// - Must have dimensions larger than 1x1
	// - Usually has a high z-index
	if w.IsFloating() && w.Width > 1 && w.Height > 1 {
		// Additional heuristics could be added here
		return true
	}

	return false
}
