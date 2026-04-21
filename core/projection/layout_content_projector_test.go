package projection

import (
	"testing"

	"nvim-gui/core/model"
)

func TestWindowBelongsToCurrentTab(t *testing.T) {
	win := &model.WindowState{ID: 1, TabID: 2}
	if !windowBelongsToCurrentTab(win, 2) {
		t.Fatal("expected window to belong to current tab")
	}
	if windowBelongsToCurrentTab(win, 3) {
		t.Fatal("expected window to be filtered from different tab")
	}
	if !windowBelongsToCurrentTab(win, 0) {
		t.Fatal("expected currentTab=0 to allow window")
	}
	if windowBelongsToCurrentTab(nil, 2) {
		t.Fatal("expected nil window to be rejected")
	}
}

func TestMapLayoutWindows_FiltersInactiveTabs(t *testing.T) {
	s := &model.ScreenState{
		CurrentTab: 2,
		Mode:       "normal",
		Windows: map[int]*model.WindowState{
			1: {ID: 1, TabID: 1, Width: 10, Height: 10},
			2: {ID: 2, TabID: 2, Width: 20, Height: 20},
		},
	}

	windows := mapLayoutWindows(s)
	if len(windows) != 1 {
		t.Fatalf("expected 1 layout window, got %d", len(windows))
	}
	if windows[0].ID != 2 {
		t.Fatalf("expected active-tab window id 2, got %d", windows[0].ID)
	}
}
