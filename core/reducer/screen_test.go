package reducer

import (
	"testing"

	"nvim-gui/core/events"
	"nvim-gui/core/model"
	fileexplorer "nvim-gui/features/file-explorer"
)

func TestReducer_CurrentTabChanged_BackfillsUnknownWindowTabsOnFirstTabEvent(t *testing.T) {
	r := NewReducer(fileexplorer.NewFileExplorer(nil))
	state := model.NewAppState()
	s := &state.Editor.Screen
	s.Windows[99] = &model.WindowState{ID: 99, GridID: 1, TabID: 0}

	r.Apply(state, events.Event{
		Name:    events.EventCurrentTabChanged,
		Payload: events.CurrentTabChanged{TabID: 2},
	})

	if s.CurrentTab != 2 {
		t.Fatalf("expected current tab 2, got %d", s.CurrentTab)
	}
	if s.Windows[99].TabID != 2 {
		t.Fatalf("expected window tab backfilled to 2, got %d", s.Windows[99].TabID)
	}
}

func TestReducer_CurrentTabChanged_NoMutationWhenTabUnchanged(t *testing.T) {
	r := NewReducer(fileexplorer.NewFileExplorer(nil))
	state := model.NewAppState()
	s := &state.Editor.Screen
	s.CurrentTab = 4
	s.ActiveWindow = 9
	s.Grids[1] = &model.GridState{ID: 1}
	s.Windows[9] = &model.WindowState{ID: 9, TabID: 4}

	r.Apply(state, events.Event{
		Name:    events.EventCurrentTabChanged,
		Payload: events.CurrentTabChanged{TabID: 4},
	})

	if s.ActiveWindow != 9 {
		t.Fatalf("expected state unchanged, active window=%d", s.ActiveWindow)
	}
	if len(s.Grids) != 1 || len(s.Windows) != 1 {
		t.Fatalf("expected state unchanged, grids=%d windows=%d", len(s.Grids), len(s.Windows))
	}
}

func TestReducer_WinPos_AssignsCurrentTabToWindow(t *testing.T) {
	r := NewReducer(fileexplorer.NewFileExplorer(nil))
	state := model.NewAppState()
	state.Editor.Screen.CurrentTab = 7

	r.Apply(state, events.Event{
		Name: events.EventWinPos,
		Payload: events.WindowPosition{
			WindowID: 11,
			GridID:   3,
			Row:      0,
			Col:      0,
			Width:    80,
			Height:   24,
		},
	})

	win := state.Editor.Screen.Windows[11]
	if win == nil {
		t.Fatal("expected window created")
	}
	if win.TabID != 7 {
		t.Fatalf("expected window tab 7, got %d", win.TabID)
	}
}
