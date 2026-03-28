package model

type AppState struct {
	Editor   EditorState
	Features FeatureState
}

type EditorState struct {
	Screen ScreenState
}

type FeatureState struct {
	Markdown     MarkdownState
	Completion   CompletionState
	FileExplorer FileExplorerState
	Picker       PickerState
}

type ScreenState struct {
	Width        int
	Height       int
	Mode         string
	ActiveWindow int
	Grids        map[int]*GridState
	Windows      map[int]*WindowState
	GridToWindow map[int]int
	Viewport     ViewportState
	Layout       LayoutState
	Cmdline      CmdlineState
	Highlights   HighlightState
}

type GridState struct {
	ID        int
	Width     int
	Height    int
	TopLine   int
	CursorRow int
	CursorCol int

	// Proper cell storage for the rendering pipeline.
	// Cells[row][col] is a *Cell with resolved highlight classes.
	Cells     [][]*Cell
	DirtyRows []bool
}

type WindowState struct {
	ID         int
	GridID     int
	Type       string
	StartRow   int
	StartCol   int
	Width      int
	Height     int
	Hidden     bool
	Focusable  bool
	Anchor     string
	AnchorGrid int
	ZIndex     int
	Filetype   string
	Filepath   string
	Dirty      bool // set by reducer when window content may have changed
}

type ViewportState struct {
	TopLine     int
	BottomLine  int
	CursorLine  int
	LineCount   int
	ScrollDelta float64
	Margins     [4]int
}

type LayoutState struct {
	Cols           string
	Rows           string
	ActiveWindowID int
}

type CmdlineState struct {
	Visible bool
	Pos     int
	Level   int
	Firstc  string
	Prompt  string
	Indent  int
	Chunks  []any
}

type HighlightState struct {
	DefaultFG   int
	DefaultBG   int
	DefaultSP   int
	Definitions [][]any
}

type MarkdownState struct{}
type CompletionState struct{}
type FileExplorerState struct{}
type PickerState struct{}

// EnsureCells allocates or resizes the Cells and DirtyRows arrays to match Width x Height.
// Existing cell data is preserved where possible.
func (g *GridState) EnsureCells() {
	if len(g.Cells) == g.Height && (g.Height == 0 || len(g.Cells[0]) == g.Width) {
		return
	}
	oldCells := g.Cells
	g.Cells = make([][]*Cell, g.Height)
	g.DirtyRows = make([]bool, g.Height)
	for row := 0; row < g.Height; row++ {
		g.Cells[row] = make([]*Cell, g.Width)
		for col := 0; col < g.Width; col++ {
			if row < len(oldCells) && col < len(oldCells[row]) && oldCells[row][col] != nil {
				g.Cells[row][col] = oldCells[row][col]
			} else {
				g.Cells[row][col] = &Cell{Char: " ", Highlight: 0, Classes: map[string]bool{}}
			}
		}
		g.DirtyRows[row] = true
	}
}

// MarkRowDirty marks a specific row as needing re-rendering.
func (g *GridState) MarkRowDirty(row int) {
	if row >= 0 && row < len(g.DirtyRows) {
		g.DirtyRows[row] = true
	}
}

// MarkAllDirty marks every row as dirty.
func (g *GridState) MarkAllDirty() {
	for i := range g.DirtyRows {
		g.DirtyRows[i] = true
	}
}

func NewAppState() *AppState {
	return &AppState{
		Editor: EditorState{
			Screen: ScreenState{
				Mode:         "normal",
				Grids:        make(map[int]*GridState),
				Windows:      make(map[int]*WindowState),
				GridToWindow: make(map[int]int),
				Layout: LayoutState{
					Cols: "1fr",
					Rows: "1fr",
				},
			},
		},
	}
}
