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
	Lines     map[int]LineState
}

type LineState struct {
	Row   int
	Col   int
	Cells []any
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
