package neovim

import (
	"context"
	"fmt"
	"log"
	"strings"

	"nvim-gui/utils"

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
	nvim_cmd := fmt.Sprintf("return require('nvim-gui.foo').attach_buffer(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}

	nvim_cmd = fmt.Sprintf("return require('nvim-gui.foo').listen_for_visual_selection_change(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	nvim_cmd = fmt.Sprintf("return require('nvim-gui.foo').listen_for_cursor_move(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	nvim_cmd = fmt.Sprintf("return require('nvim-gui.foo').listen_for_mode_change(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}

	v.RegisterHandler("nvim-gui-buf-changed", func(v *nvim.Nvim, root NvimGuiNode) {
		current_tree = root
		
		fmt.Println("")
		fmt.Println("")
		fmt.Print(PrintNode(current_tree, ""))
		fmt.Println("")
		fmt.Println("")
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
func PrintNode(node NvimGuiNode, indent string) string {
	var result strings.Builder

	// Write basic node information
	result.WriteString(fmt.Sprintf("%sNode ID: %s\n", indent, node.Id))
	result.WriteString(fmt.Sprintf("%sPosition: (%d:%d) -> (%d:%d)\n",
		indent, node.StartRow, node.StartCol, node.EndRow, node.EndCol))

	// Write text content with special formatting for linebreaks and spaces
	if node.LineBreak {
		result.WriteString(fmt.Sprintf("%sType: LineBreak\n", indent))
	} else if node.Space {
		result.WriteString(fmt.Sprintf("%sType: Space\n", indent))
	}

	if node.Text != "" {
		// Replace invisible characters with visible representations
		text := strings.ReplaceAll(node.Text, "\n", "⏎")
		text = strings.ReplaceAll(text, " ", "␣")
		result.WriteString(fmt.Sprintf("%sText: '%s'\n", indent, text))
	}

	if node.HlGroup != "" {
		result.WriteString(fmt.Sprintf("%sHighlight: %s\n", indent, node.HlGroup))
	}

	// Print children recursively
	if len(node.Children) > 0 {
		result.WriteString(fmt.Sprintf("%sChildren (%d):\n", indent, len(node.Children)))
		for i, child := range node.Children {
			result.WriteString(fmt.Sprintf("%s└─── Child %d:\n", indent, i+1))
			result.WriteString(PrintNode(child, indent+"    "))
		}
	}

	return result.String()
}
