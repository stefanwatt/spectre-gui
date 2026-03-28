package rendering

import (
	"testing"
)

func TestFromCell(t *testing.T) {
	cell := &Cell{
		Char:      "x",
		Highlight: 5,
		Classes:   map[string]bool{"fg-1": true, "font-bold": true},
	}
	token := fromCell(cell)

	if token.Text != "x" {
		t.Errorf("Text = %q, want %q", token.Text, "x")
	}
	if token.Highlight != 5 {
		t.Errorf("Highlight = %d, want 5", token.Highlight)
	}
	// ClassesToString sorts alphabetically: "fg-1 font-bold"
	want := "fg-1 font-bold"
	if token.Classes != want {
		t.Errorf("Classes = %q, want %q", token.Classes, want)
	}
}

func makeCell(char string, hl int, classes map[string]bool) *Cell {
	if classes == nil {
		classes = make(map[string]bool)
	}
	return &Cell{Char: char, Highlight: hl, Classes: classes}
}

func TestOptimizeRow_MergesSameHl(t *testing.T) {
	s := &Screen{}
	cells := []*Cell{
		makeCell("a", 1, map[string]bool{"fg-1": true}),
		makeCell("b", 1, map[string]bool{"fg-1": true}),
		makeCell("c", 1, map[string]bool{"fg-1": true}),
	}
	cursor := struct {
		Row int
		Col int
	}{Row: -1, Col: -1} // cursor not on this row

	result := s.optimizeRow(cells, 0, cursor)

	if len(result) != 1 {
		t.Fatalf("got %d tokens, want 1", len(result))
	}
	if result[0].Char != "abc" {
		t.Errorf("merged text = %q, want %q", result[0].Char, "abc")
	}
}

func TestOptimizeRow_SplitsDifferentHl(t *testing.T) {
	s := &Screen{}
	cells := []*Cell{
		makeCell("a", 1, map[string]bool{"fg-1": true}),
		makeCell("b", 2, map[string]bool{"fg-2": true}),
		makeCell("c", 1, map[string]bool{"fg-1": true}),
	}
	cursor := struct {
		Row int
		Col int
	}{Row: -1, Col: -1}

	result := s.optimizeRow(cells, 0, cursor)

	if len(result) != 3 {
		t.Fatalf("got %d tokens, want 3", len(result))
	}
	if result[0].Char != "a" || result[1].Char != "b" || result[2].Char != "c" {
		t.Errorf("tokens = [%q, %q, %q], want [a, b, c]",
			result[0].Char, result[1].Char, result[2].Char)
	}
}

func TestOptimizeRow_CursorInjection(t *testing.T) {
	s := &Screen{}
	cells := []*Cell{
		makeCell("a", 1, map[string]bool{"fg-1": true}),
		makeCell("b", 1, map[string]bool{"fg-1": true}),
		makeCell("c", 1, map[string]bool{"fg-1": true}),
	}
	cursor := struct {
		Row int
		Col int
	}{Row: 0, Col: 1} // cursor on cell "b"

	result := s.optimizeRow(cells, 0, cursor)

	// Should have: "a", cursor "b", "c"
	if len(result) != 3 {
		t.Fatalf("got %d tokens, want 3", len(result))
	}

	// The cursor cell should have the "cursor" class
	cursorClasses := result[1].ClassesToString()
	found := false
	for class := range result[1].Classes {
		if class == "cursor" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("cursor cell classes = %q, missing 'cursor'", cursorClasses)
	}
	if result[1].Char != "b" {
		t.Errorf("cursor cell char = %q, want %q", result[1].Char, "b")
	}
}

func TestOptimizeRow_CursorSplitsToken(t *testing.T) {
	s := &Screen{}
	cells := []*Cell{
		makeCell("a", 1, map[string]bool{"fg-1": true}),
		makeCell("b", 1, map[string]bool{"fg-1": true}),
		makeCell("c", 1, map[string]bool{"fg-1": true}),
		makeCell("d", 1, map[string]bool{"fg-1": true}),
	}
	cursor := struct {
		Row int
		Col int
	}{Row: 0, Col: 2} // cursor on "c"

	result := s.optimizeRow(cells, 0, cursor)

	// "ab" merged, cursor "c", "d"
	if len(result) != 3 {
		t.Fatalf("got %d tokens, want 3: %v", len(result), result)
	}
	if result[0].Char != "ab" {
		t.Errorf("first token = %q, want %q", result[0].Char, "ab")
	}
	if result[1].Char != "c" {
		t.Errorf("cursor token = %q, want %q", result[1].Char, "c")
	}
	if _, ok := result[1].Classes["cursor"]; !ok {
		t.Error("cursor cell missing 'cursor' class")
	}
	if result[2].Char != "d" {
		t.Errorf("last token = %q, want %q", result[2].Char, "d")
	}
}

func TestOptimizeRow_TrailingWhitespace(t *testing.T) {
	s := &Screen{}
	cells := []*Cell{
		makeCell("h", 1, map[string]bool{"fg-1": true}),
		makeCell("i", 1, map[string]bool{"fg-1": true}),
		makeCell(" ", 1, map[string]bool{"fg-1": true}),
		makeCell(" ", 1, map[string]bool{"fg-1": true}),
	}
	cursor := struct {
		Row int
		Col int
	}{Row: -1, Col: -1}

	result := s.optimizeRow(cells, 0, cursor)

	if len(result) != 1 {
		t.Fatalf("got %d tokens, want 1", len(result))
	}
	if result[0].Char != "hi" {
		t.Errorf("text = %q, want %q (trailing spaces trimmed)", result[0].Char, "hi")
	}
}

func TestOptimizeRow_EmptyRow(t *testing.T) {
	s := &Screen{}
	cursor := struct {
		Row int
		Col int
	}{Row: -1, Col: -1}

	result := s.optimizeRow([]*Cell{}, 0, cursor)

	if len(result) != 0 {
		t.Errorf("got %d tokens, want 0", len(result))
	}
}

func TestSplitTokensIntoCells_BasicTable(t *testing.T) {
	tokens := []*Token{
		{Text: "| Name | Age | City |", Classes: "fg-1", Highlight: 1},
	}
	cells := splitTokensIntoCells(tokens)

	if len(cells) != 3 {
		t.Fatalf("got %d cells, want 3", len(cells))
	}
	if cells[0][0].Text != "Name" {
		t.Errorf("cell 0 text = %q, want %q", cells[0][0].Text, "Name")
	}
	if cells[1][0].Text != "Age" {
		t.Errorf("cell 1 text = %q, want %q", cells[1][0].Text, "Age")
	}
	if cells[2][0].Text != "City" {
		t.Errorf("cell 2 text = %q, want %q", cells[2][0].Text, "City")
	}
}

func TestSplitTokensIntoCells_MultipleTokensPerCell(t *testing.T) {
	// Simulates "| " with hl=1, "Name" with hl=2, " | " with hl=1, "Age" with hl=2, " |"
	tokens := []*Token{
		{Text: "| ", Classes: "fg-1", Highlight: 1},
		{Text: "Name", Classes: "fg-2", Highlight: 2},
		{Text: " | ", Classes: "fg-1", Highlight: 1},
		{Text: "Age", Classes: "fg-2", Highlight: 2},
		{Text: " |", Classes: "fg-1", Highlight: 1},
	}
	cells := splitTokensIntoCells(tokens)

	if len(cells) != 2 {
		t.Fatalf("got %d cells, want 2", len(cells))
	}
	if cells[0][0].Text != "Name" {
		t.Errorf("cell 0 text = %q, want %q", cells[0][0].Text, "Name")
	}
	if cells[1][0].Text != "Age" {
		t.Errorf("cell 1 text = %q, want %q", cells[1][0].Text, "Age")
	}
}

func TestSplitTokensIntoCells_EmptyInput(t *testing.T) {
	cells := splitTokensIntoCells([]*Token{})
	if len(cells) != 0 {
		t.Errorf("got %d cells, want 0", len(cells))
	}
}

func TestSplitTokensIntoCells_BoxDrawing(t *testing.T) {
	// Box-drawing pipes from render-markdown.nvim
	tokens := []*Token{
		{Text: "│ Name  │ Age │ City     │", Classes: "fg-1", Highlight: 1},
	}
	cells := splitTokensIntoCells(tokens)

	if len(cells) != 3 {
		t.Fatalf("got %d cells, want 3", len(cells))
	}
	if cells[0][0].Text != "Name" {
		t.Errorf("cell 0 text = %q, want %q", cells[0][0].Text, "Name")
	}
	if cells[1][0].Text != "Age" {
		t.Errorf("cell 1 text = %q, want %q", cells[1][0].Text, "Age")
	}
	if cells[2][0].Text != "City" {
		t.Errorf("cell 2 text = %q, want %q", cells[2][0].Text, "City")
	}
}

func TestIsSeparatorRow(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"| --- | --- | --- |", true},
		{"| :--- | :---: | ---: |", true},
		{"| Name | Age |", false},
		{"just text", false},
		{"├───────┼─────┼──────────┤", true},
	}
	for _, tt := range tests {
		tokens := []*Token{{Text: tt.text}}
		got := isSeparatorRow(tokens)
		if got != tt.want {
			t.Errorf("isSeparatorRow(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestBuildTableRowOpts_Header(t *testing.T) {
	meta := &TableMeta{StartLine: 5, EndLine: 9, Alignments: []string{"left", "center"}}
	tokens := []*Token{{Text: "| Name | Age |"}}

	opts := buildTableRowOpts(meta, 5, tokens)

	if opts.RowType != "header" {
		t.Errorf("RowType = %q, want %q", opts.RowType, "header")
	}
	if opts.TableID != 5 {
		t.Errorf("TableID = %d, want 5", opts.TableID)
	}
	if len(opts.Alignments) != 2 {
		t.Fatalf("Alignments len = %d, want 2", len(opts.Alignments))
	}
}

func TestBuildTableRowOpts_Separator(t *testing.T) {
	meta := &TableMeta{StartLine: 5, EndLine: 9, Alignments: []string{"left"}}
	tokens := []*Token{{Text: "| --- | --- |"}}

	opts := buildTableRowOpts(meta, 6, tokens)

	if opts.RowType != "separator" {
		t.Errorf("RowType = %q, want %q", opts.RowType, "separator")
	}
}

func TestBuildTableRowOpts_Data(t *testing.T) {
	meta := &TableMeta{StartLine: 5, EndLine: 9, Alignments: []string{"left"}}
	tokens := []*Token{{Text: "| Alice | 30 |"}}

	opts := buildTableRowOpts(meta, 7, tokens)

	if opts.RowType != "data" {
		t.Errorf("RowType = %q, want %q", opts.RowType, "data")
	}
}
