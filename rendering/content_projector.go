package rendering

type ContentInput struct {
	WindowID   int
	Filetype   string
	BufNr      int
	CursorLine int
	Grid       *GridData
	Meta       MarkdownMetaProvider
}

type ContentOutput struct {
	WindowID int          `json:"winId"`
	Content  []ContentRow `json:"updatedContent"`
}

func BuildContentPayload(input ContentInput) ContentOutput {
	if input.Grid == nil {
		return ContentOutput{WindowID: input.WindowID, Content: []ContentRow{}}
	}
	rows := OptimizeGrid(input.Grid, input.Filetype, input.BufNr, input.CursorLine, input.Meta)
	return ContentOutput{WindowID: input.WindowID, Content: rows}
}
