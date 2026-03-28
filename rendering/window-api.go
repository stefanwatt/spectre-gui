package rendering

type WindowAPI struct {
	ID                  int     `json:"id"`
	Type                string  `json:"type"`
	Width               float64 `json:"width"`  // in percent of screen
	Height              float64 `json:"height"` // in percent of screen
	ColStart            int     `json:"colStart"`
	ColEnd              int     `json:"colEnd"`
	RowStart            int     `json:"rowStart"`
	RowEnd              int     `json:"rowEnd"`
	LineNumbers         bool    `json:"lineNumbers"`
	RelativeLineNumbers bool    `json:"relativeLineNumbers"`
	Filetype            string  `json:"filetype"`
	Filepath            string  `json:"filepath"`
	Mode                string  `json:"mode"`
	Cursor              *Cursor `json:"cursor"`
	ColorColumns        []int   `json:"colorColumns"`
	ColorColumnColor    string  `json:"colorColumnColor"`
	CursorLineColor     string  `json:"cursorLineColor"`
}

// Equal compares two WindowAPI structs and returns true if they are equal.
func (w *WindowAPI) Equal(other *WindowAPI) bool {
	if other == nil {
		return false
	}

	return w.ID == other.ID &&
		w.Type == other.Type &&
		w.Width == other.Width &&
		w.Height == other.Height &&
		w.ColStart == other.ColStart &&
		w.ColEnd == other.ColEnd &&
		w.RowStart == other.RowStart &&
		w.RowEnd == other.RowEnd &&
		w.LineNumbers == other.LineNumbers &&
		w.RelativeLineNumbers == other.RelativeLineNumbers &&
		w.Filetype == other.Filetype &&
		w.Filepath == other.Filepath &&
		w.Mode == other.Mode &&
		w.ColorColumnColor == other.ColorColumnColor &&
		w.CursorLineColor == other.CursorLineColor
}
