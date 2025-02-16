package neovim

import (
	"context"
	"spectre-gui/utils"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type NvimGuiNode struct {
	Id        string        `msgpack:"id" json:"id"`
	Text      string        `msgpack:"text" json:"text"`
	HlGroup   string        `msgpack:"hl_group" json:"hl_group"`
	StartTop  uint64        `msgpack:"start_top" json:"-"`
	EndTop    uint64        `msgpack:"end_top" json:"-"`
	StartLeft uint64        `msgpack:"start_left" json:"-"`
	EndLeft   uint64        `msgpack:"end_left" json:"-"`
	StartRow  uint64        `msgpack:"start_row" json:"-"`
	StartCol  uint64        `msgpack:"start_col" json:"-"`
	EndRow    uint64        `msgpack:"end_row" json:"-"`
	EndCol    uint64        `msgpack:"end_col" json:"-"`
	Root      bool          `msgpack:"root" json:"-"`
	LineBreak bool          `msgpack:"line_break" json:"line_break"`
	Space     bool          `msgpack:"space" json:"space"`
	Children  []NvimGuiNode `msgpack:"children" json:"children"`
}

func OnBufChanged(ctx context.Context, root NvimGuiNode) {
	utils.LogTimeSinceLast("update buf lines")
	Runtime.EventsEmit(ctx, "buf-lines-changed", root)
}

func SplitNodeAtCursor(root NvimGuiNode, cursorRow uint64, cursorCol uint64) NvimGuiNode {
	return processNode(root, cursorRow, cursorCol)
}

func isInNode(node NvimGuiNode, cursorRow uint64, cursorCol uint64) bool {
	if node.StartRow > cursorRow || node.EndRow < cursorRow {
		return false
	}
	if node.StartRow == cursorRow && node.StartCol > cursorCol {
		return false
	}
	if node.EndRow == cursorRow && node.EndCol <= cursorCol {
		return false
	}
	return true
}

func processNode(node NvimGuiNode, cursorRow uint64, cursorCol uint64) NvimGuiNode {
	if !isInNode(node, cursorRow, cursorCol) {
		return node
	}
	// If this is a leaf node containing the cursor
	if len(node.Children) == 0 && isInNode(node, cursorRow, cursorCol) {
		before, cursor, after := splitTextAtCursor(node, cursorRow, cursorCol)
		return NvimGuiNode{
			Id:       node.Id + "_container",
			HlGroup:  node.HlGroup,
			Root:     node.Root,
			Children: []NvimGuiNode{before, cursor, after},
		}
	}

	// Process children if this is not a leaf node
	newChildren := make([]NvimGuiNode, len(node.Children))
	for i, child := range node.Children {
		newChildren[i] = processNode(child, cursorRow, cursorCol)
	}
	node.Children = newChildren
	return node
}

func splitTextAtCursor(node NvimGuiNode, cursorRow uint64, cursorCol uint64) (before, cursor, after NvimGuiNode) {
	if node.StartRow == node.EndRow {
		// Cursor is in a single-line node
		relativeCol := cursorCol - node.StartCol
		beforeText := node.Text[:relativeCol]
		cursorText := string([]rune(node.Text)[relativeCol : relativeCol+1])
		afterText := node.Text[relativeCol+1:]

		before = NvimGuiNode{
			Id:        node.Id + "_before",
			Text:      beforeText,
			HlGroup:   node.HlGroup,
			StartRow:  node.StartRow,
			StartCol:  node.StartCol,
			EndRow:    node.StartRow,
			EndCol:    cursorCol,
			LineBreak: false,
			Space:     node.Space,
		}

		cursor = NvimGuiNode{
			Id:        node.Id + "_cursor",
			Text:      cursorText,
			HlGroup:   "cursor",
			StartRow:  cursorRow,
			StartCol:  cursorCol,
			EndRow:    cursorRow,
			EndCol:    cursorCol + 1,
			LineBreak: false,
			Space:     false,
		}

		after = NvimGuiNode{
			Id:        node.Id + "_after",
			Text:      afterText,
			HlGroup:   node.HlGroup,
			StartRow:  node.StartRow,
			StartCol:  cursorCol + 1,
			EndRow:    node.EndRow,
			EndCol:    node.EndCol,
			LineBreak: node.LineBreak,
			Space:     node.Space,
		}
	}
	return
}
