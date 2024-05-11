package match

import "testing"

func TestMapReplacementTextPreserveCase(t *testing.T) {
	actual := map_replacement_text_preserve_case("foo", "bar")

	expected := "bar"
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

	actual = map_replacement_text_preserve_case("Foo", "bar")
	expected = "Bar"
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}

	actual = map_replacement_text_preserve_case("FOO", "bar")
	expected = "BAR"
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}
