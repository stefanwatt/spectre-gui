package neovim

import (
	"context"
	"fmt"
	"log"

	"spectre-gui/utils"

	"github.com/neovim/go-client/nvim"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var current_tree NvimGuiNode

func StartListening(servername string, ctx context.Context) {
	v, err := nvim.Dial(servername)
	if err != nil {
		log.Println(err)
		return
	}
	defer v.Close()
	var result string
	nvim_cmd := fmt.Sprintf("return require('config.nvim-gui').attach_buffer(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}

	nvim_cmd = fmt.Sprintf("return require('config.nvim-gui').listen_for_visual_selection_change(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	nvim_cmd = fmt.Sprintf("return require('config.nvim-gui').listen_for_cursor_move(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	nvim_cmd = fmt.Sprintf("return require('config.nvim-gui').listen_for_mode_change(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}

	v.RegisterHandler("nvim-gui-buf-changed", func(v *nvim.Nvim, root NvimGuiNode) {
		current_tree = root
		OnBufChanged(ctx, root)
	})

	v.RegisterHandler("nvim-gui-cursor-moved", func(v *nvim.Nvim, cursor_move_event CursorMoveEvent) {
		row := cursor_move_event.Row
		col := cursor_move_event.Col
		updated_tree := SplitNodeAtCursor(current_tree, row, col)
		OnBufChanged(ctx, updated_tree)
		UpdateCursor(ctx, cursor_move_event)
	})

	v.RegisterHandler("nvim-gui-mode-changed", func(v *nvim.Nvim, args []string) {
		mode := args[0]
		UpdateSelection(ctx, mode)
	})

	v.RegisterHandler("nvim-gui-visual-selection-changed", func(v *nvim.Nvim, selection_range NvimRange) {
		Runtime.EventsEmit(ctx, "visual-selection-changed", selection_range)
	})

	if err := v.Serve(); err != nil {
		log.Fatal(err)
	}
	log.Println("listening terminating")
}
