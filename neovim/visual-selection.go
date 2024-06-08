package neovim

import (
	"context"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func UpdateSelection(ctx context.Context, mode string) {

	Runtime.EventsEmit(ctx, "mode-changed", mode)
	if mode != "v" && mode != "V" {
		Runtime.EventsEmit(ctx, "visual-selection-changed", NvimRange{
			StartRow: 9999999,
			EndRow:   9999999,
			StartCol: 0,
			EndCol:   0,
		})
	}
}
