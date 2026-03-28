package rendering

type MarkdownOpts struct {
	QuoteLevel   int            `json:"quoteLevel"`
	HeadingLevel int            `json:"headingLevel,omitempty"`
	Table        *TableRowOpts  `json:"table,omitempty"`
	Image        *ImageOpts     `json:"image,omitempty"`
	Task         *TaskOpts      `json:"task,omitempty"`
	CodeBlock    *CodeBlockOpts `json:"codeBlock,omitempty"`
}

type TableMeta struct {
	StartLine  int
	EndLine    int
	Alignments []string
}

type TableRowOpts struct {
	TableID    int        `json:"tableId"`
	RowType    string     `json:"rowType"`
	Cells      [][]*Token `json:"cells"`
	Alignments []string   `json:"alignments"`
}

type ImageMeta struct {
	Line    int
	URL     string
	AltText string
}

type ImageOpts struct {
	URL     string `json:"url"`
	AltText string `json:"altText"`
}

type TaskOpts struct {
	Checked bool `json:"checked"`
}

type CodeBlockMeta struct {
	StartLine int
	EndLine   int
}

func buildTableRowOpts(meta *TableMeta, bufferLine int, tokens []*Token) *TableRowOpts {
	if isBorderRow(tokens) || isSeparatorRow(tokens) {
		return &TableRowOpts{
			TableID:    meta.StartLine,
			RowType:    "separator",
			Cells:      nil,
			Alignments: meta.Alignments,
		}
	}

	cells := splitTokensIntoCells(tokens)
	rowType := "data"
	if bufferLine == meta.StartLine {
		rowType = "header"
	}

	return &TableRowOpts{
		TableID:    meta.StartLine,
		RowType:    rowType,
		Cells:      cells,
		Alignments: meta.Alignments,
	}
}
