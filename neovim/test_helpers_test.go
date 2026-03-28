package neovim

import "nvim-gui/rendering"

func resetHighlightState() {
	rendering.ResetHighlightStateForTest()
}
