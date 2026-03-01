package neovim

import (
	"testing"
)

func TestFgHex_NoForeground(t *testing.T) {
	h := &Highlight{HasForeground: false, Foreground: 0xff0000}
	got := h.fgHex()
	if got != "" {
		t.Errorf("fgHex() = %q, want empty string", got)
	}
}

func TestFgHex_WithForeground(t *testing.T) {
	h := &Highlight{HasForeground: true, Foreground: 0xff0000}
	got := h.fgHex()
	want := "#ff0000"
	if got != want {
		t.Errorf("fgHex() = %q, want %q", got, want)
	}
}

func TestBgHex_NoBackground(t *testing.T) {
	h := &Highlight{HasBackground: false, Background: 0x00ff00}
	got := h.bgHex()
	if got != "" {
		t.Errorf("bgHex() = %q, want empty string", got)
	}
}

func TestBgHex_WithBackground(t *testing.T) {
	h := &Highlight{HasBackground: true, Background: 0x00ff00}
	got := h.bgHex()
	want := "#00ff00"
	if got != want {
		t.Errorf("bgHex() = %q, want %q", got, want)
	}
}

func TestGetClasses_ForegroundOnly(t *testing.T) {
	resetHighlightState()
	h := &Highlight{HasForeground: true, Foreground: 0xaabbcc}
	classes := h.getClasses()
	if len(classes) != 1 {
		t.Fatalf("getClasses() returned %d classes, want 1", len(classes))
	}
	if classes[0] != "fg-1" {
		t.Errorf("getClasses()[0] = %q, want %q", classes[0], "fg-1")
	}
}

func TestGetClasses_FgBgBoldItalic(t *testing.T) {
	resetHighlightState()
	h := &Highlight{
		HasForeground: true, Foreground: 0x112233,
		HasBackground: true, Background: 0x445566,
		Bold:   true,
		Italic: true,
	}
	classes := h.getClasses()
	if len(classes) != 4 {
		t.Fatalf("getClasses() returned %d classes, want 4: %v", len(classes), classes)
	}
	expected := []string{"fg-1", "bg-1", "font-bold", "italic"}
	for i, want := range expected {
		if classes[i] != want {
			t.Errorf("getClasses()[%d] = %q, want %q", i, classes[i], want)
		}
	}
}

func TestGetClasses_AllFontStyles(t *testing.T) {
	resetHighlightState()
	h := &Highlight{
		Bold:          true,
		Italic:        true,
		Underline:     true,
		Undercurl:     true,
		Strikethrough: true,
	}
	classes := h.getClasses()
	if len(classes) != 5 {
		t.Fatalf("getClasses() returned %d classes, want 5: %v", len(classes), classes)
	}
	expected := []string{"font-bold", "italic", "underline", "underline decoration-wavy", "line-through"}
	for i, want := range expected {
		if classes[i] != want {
			t.Errorf("getClasses()[%d] = %q, want %q", i, classes[i], want)
		}
	}
}

func TestGetClasses_NoAttributes(t *testing.T) {
	resetHighlightState()
	h := &Highlight{}
	classes := h.getClasses()
	if len(classes) != 0 {
		t.Errorf("getClasses() returned %d classes, want 0: %v", len(classes), classes)
	}
}

func TestGetClasses_Idempotent(t *testing.T) {
	resetHighlightState()
	h := &Highlight{HasForeground: true, Foreground: 0xddeeff}

	classes1 := h.getClasses()
	classes2 := h.getClasses()

	if len(classes1) != 1 || len(classes2) != 1 {
		t.Fatalf("expected 1 class each, got %d and %d", len(classes1), len(classes2))
	}
	if classes1[0] != classes2[0] {
		t.Errorf("same highlight produced different classes: %q vs %q", classes1[0], classes2[0])
	}
	if classes1[0] != "fg-1" {
		t.Errorf("expected fg-1, got %q", classes1[0])
	}
}

func TestMapClassesString(t *testing.T) {
	got := mapClassesString([]string{"fg-1", "bg-2"})
	want := "fg-1-bg-2"
	if got != want {
		t.Errorf("mapClassesString() = %q, want %q", got, want)
	}
}
