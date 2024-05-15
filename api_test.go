package main

import (
	"context"
	"os"
	"testing"
)

func setup_dir() {
	os.RemoveAll("/tmp/spectre-gui-test")
	os.MkdirAll("/tmp/spectre-gui-test", 0o755)
	os.Create("/tmp/spectre-gui-test/test.txt")
	sample_text := `// some text ...  foo
// sht Foo shtisena
// oeshnt osenfooht
// shtsth FOO shiifupfdr
// sht foo shtioensh`
	os.WriteFile("/tmp/spectre-gui-test/test.txt", []byte(sample_text), 0o644)
}

var app *App

func TestSearch(t *testing.T) {
	setup_dir()
	app = NewApp()
	ctx := context.Background()
	app.startup(ctx)

	t.Run("Search for 'foo' without flags", TestSearchSimple)
	t.Run("Search for 'foo' preserving case", TestSearchPreserveCase)
}

func TestSearchSimple(t *testing.T) {
	result := app.Search(
		"foo",
		"",
		"/tmp/spectre-gui-test",
		"",
		"",
		false,
		false,
		false,
		false,
	)

	grouped_matches := result.GroupedMatches
	if len(grouped_matches) != 1 {
		t.Errorf("Expected results to have 1 file, got %d", len(grouped_matches))
	}
	if len(grouped_matches[0].Matches) != 5 {
		t.Errorf("Expected file %s to have 5 matches, got %d", grouped_matches[0].Path, len(grouped_matches[0].Matches))
	}
}

func TestSearchPreserveCase(t *testing.T) {
	result := app.Search(
		"foo",
		"bar",
		"/tmp/spectre-gui-test",
		"",
		"",
		false,
		false,
		false,
		true,
	)

	matches := result.GroupedMatches[0].Matches
	expected := "bar"
	actual := matches[0].ReplacementText
	if actual != expected {
		t.Errorf("Expected replacement text to be '%s', got '%s'", expected, actual)
	}
	expected = "Bar"
	actual = matches[1].ReplacementText
	if actual != expected {
		t.Errorf("Expected replacement text to be '%s', got '%s'", expected, actual)
	}
	expected = "bar"
	actual = matches[2].ReplacementText
	if actual != expected {
		t.Errorf("Expected replacement text to be '%s', got '%s'", expected, actual)
	}
	expected = "BAR"
	actual = matches[3].ReplacementText
	if actual != expected {
		t.Errorf("Expected replacement text to be '%s', got '%s'", expected, actual)
	}
	expected = "bar"
	actual = matches[4].ReplacementText
	if actual != expected {
		t.Errorf("Expected replacement text to be '%s', got '%s'", expected, actual)
	}
}
