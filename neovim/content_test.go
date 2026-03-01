package neovim

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
