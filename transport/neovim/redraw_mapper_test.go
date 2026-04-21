package neovimtransport

import (
	"testing"

	"nvim-gui/core/events"

	"github.com/neovim/go-client/nvim"
)

func TestMapRedrawBatch_PrioritizesTablineUpdate(t *testing.T) {
	updates := [][]interface{}{
		{"grid_resize", []interface{}{1, 80, 24}},
		{"tabline_update", []interface{}{nvim.Tabpage(7), []interface{}{}, nvim.Buffer(1), []interface{}{}}},
		{"flush"},
	}

	mapped := MapRedrawBatch(updates)
	if len(mapped) != 3 {
		t.Fatalf("expected 3 events, got %d", len(mapped))
	}
	if mapped[0].Name != events.EventCurrentTabChanged {
		t.Fatalf("expected first event %q, got %q", events.EventCurrentTabChanged, mapped[0].Name)
	}
	payload, ok := mapped[0].Payload.(events.CurrentTabChanged)
	if !ok {
		t.Fatalf("expected CurrentTabChanged payload, got %T", mapped[0].Payload)
	}
	if payload.TabID != 7 {
		t.Fatalf("expected tab id 7, got %d", payload.TabID)
	}
}

func TestMapRedrawBatch_MapsTablineUpdate(t *testing.T) {
	updates := [][]interface{}{
		{"tabline_update", []interface{}{nvim.Tabpage(3), []interface{}{}, nvim.Buffer(9), []interface{}{}}},
	}

	mapped := MapRedrawBatch(updates)
	if len(mapped) != 1 {
		t.Fatalf("expected 1 event, got %d", len(mapped))
	}
	if mapped[0].Name != events.EventCurrentTabChanged {
		t.Fatalf("expected %q, got %q", events.EventCurrentTabChanged, mapped[0].Name)
	}
	payload := mapped[0].Payload.(events.CurrentTabChanged)
	if payload.TabID != 3 {
		t.Fatalf("expected tab id 3, got %d", payload.TabID)
	}
}
