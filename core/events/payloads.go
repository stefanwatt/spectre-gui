package events

type GridResize struct {
	GridID int
	Width  int
	Height int
}

type GridLine struct {
	GridID int
	Row    int
	Col    int
	Cells  []any
}

type GridClear struct {
	GridID int
}

type GridScroll struct {
	GridID int
	Top    int
	Bottom int
	Left   int
	Right  int
	Rows   int
	Cols   int
}

type GridCursorGoto struct {
	GridID int
	Row    int
	Col    int
}

type WindowPosition struct {
	GridID   int
	WindowID int
	Row      int
	Col      int
	Width    int
	Height   int
}

type FloatingWindowPosition struct {
	GridID     int
	WindowID   int
	Anchor     string
	AnchorGrid int
	AnchorRow  float64
	AnchorCol  float64
	Focusable  bool
	ZIndex     int
}

type WindowViewport struct {
	GridID      int
	TopLine     int
	BottomLine  int
	CursorLine  int
	LineCount   int
	ScrollDelta float64
}

type WindowViewportMargins struct {
	GridID int
	Top    int
	Bottom int
	Left   int
	Right  int
}

type ModeChange struct {
	Mode string
}

type CmdlineShow struct {
	Chunks []any
	Pos    int
	Firstc string
	Prompt string
	Indent int
}

type CmdlinePos struct {
	Pos   int
	Level int
}

type CmdlineHide struct{}

type DefaultColorsSet struct {
	FG int
	BG int
	SP int
}

type HighlightAttrDefine struct {
	Args []any
}

type BufEnter struct {
	BufNr    int    `json:"bufnr" msgpack:"buf"`
	Filepath string `json:"filepath" msgpack:"file"`
	WinID    int    `json:"winId" msgpack:"winId"`
}

type WindowBufferInfo struct {
	WindowID int
	Filetype string
	Filepath string
}

type WindowOptions struct {
	WindowID    int
	LineNumbers bool
}
