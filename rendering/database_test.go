package rendering

import (
	"testing"
)

func TestAddFgColorClass_Increments(t *testing.T) {
	resetHighlightState()

	addForegroundColorClass("#aabbcc")
	addForegroundColorClass("#ddeeff")

	fgColorClassesMu.Lock()
	c1 := fgColorClasses["#aabbcc"]
	c2 := fgColorClasses["#ddeeff"]
	fgColorClassesMu.Unlock()

	if c1 != "fg-1" {
		t.Errorf("first color class = %q, want %q", c1, "fg-1")
	}
	if c2 != "fg-2" {
		t.Errorf("second color class = %q, want %q", c2, "fg-2")
	}
}

func TestAddFgColorClass_Idempotent(t *testing.T) {
	resetHighlightState()

	addForegroundColorClass("#aabbcc")
	addForegroundColorClass("#aabbcc")

	fgColorClassesMu.Lock()
	c := fgColorClasses["#aabbcc"]
	count := len(fgColorClasses)
	fgColorClassesMu.Unlock()

	if c != "fg-1" {
		t.Errorf("color class = %q, want %q", c, "fg-1")
	}
	if count != 1 {
		t.Errorf("map has %d entries, want 1", count)
	}
}

func TestAddFgColorClass_EmptyString(t *testing.T) {
	resetHighlightState()

	addForegroundColorClass("")

	fgColorClassesMu.Lock()
	count := len(fgColorClasses)
	fgColorClassesMu.Unlock()

	if count != 0 {
		t.Errorf("map has %d entries after empty string, want 0", count)
	}
}

func TestAddBgColorClass_Increments(t *testing.T) {
	resetHighlightState()

	addBackgroundColorClass("#112233")
	addBackgroundColorClass("#445566")

	bgColorClassesMu.Lock()
	c1 := bgColorClasses["#112233"]
	c2 := bgColorClasses["#445566"]
	bgColorClassesMu.Unlock()

	if c1 != "bg-1" {
		t.Errorf("first bg class = %q, want %q", c1, "bg-1")
	}
	if c2 != "bg-2" {
		t.Errorf("second bg class = %q, want %q", c2, "bg-2")
	}
}

func TestAddIdClasses(t *testing.T) {
	resetHighlightState()

	classes := []string{"fg-1", "font-bold"}
	addIdClasses(42, classes)

	idClassesMu.Lock()
	stored, exists := idClasses[42]
	idClassesMu.Unlock()

	if !exists {
		t.Fatal("idClasses[42] does not exist")
	}
	if len(stored) != 2 {
		t.Fatalf("stored %d classes, want 2", len(stored))
	}
	if stored[0] != "fg-1" || stored[1] != "font-bold" {
		t.Errorf("stored classes = %v, want [fg-1, font-bold]", stored)
	}

	// Should not overwrite
	addIdClasses(42, []string{"bg-1"})

	idClassesMu.Lock()
	stored2 := idClasses[42]
	idClassesMu.Unlock()

	if len(stored2) != 2 || stored2[0] != "fg-1" {
		t.Errorf("addIdClasses overwrote existing entry: %v", stored2)
	}
}
