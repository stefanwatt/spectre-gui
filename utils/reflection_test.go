package utils

import (
	"testing"

	"github.com/neovim/go-client/nvim"
)

func TestReflectToInt_NvimBuffer(t *testing.T) {
	got := ReflectToInt(nvim.Buffer(31))
	if got != 31 {
		t.Fatalf("expected 31, got %d", got)
	}
}

func TestReflectToInt_AliasedInt(t *testing.T) {
	type customInt int
	got := ReflectToInt(customInt(7))
	if got != 7 {
		t.Fatalf("expected 7, got %d", got)
	}
}
