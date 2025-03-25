package neovim

import (
	"context"
	"nvim-gui/utils"

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
	processed, _ := processNode(root, cursorRow, cursorCol)
	return processed
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

func processNode(node NvimGuiNode, cursorRow uint64, cursorCol uint64) (NvimGuiNode, bool) {
	if !isInNode(node, cursorRow, cursorCol) {
		return node, false
	}

	if len(node.Children) == 0 {
		before, cursor, after := splitTextAtCursor(node, cursorRow, cursorCol)

		// Build children list with only non-empty nodes
		children := make([]NvimGuiNode, 0, 3)
		if before.Id != "" {
			children = append(children, before)
		}
		children = append(children, cursor)
		if after.Id != "" {
			children = append(children, after)
		}

		return NvimGuiNode{
			Id:        node.Id + "_container",
			HlGroup:   node.HlGroup,
			Root:      node.Root,
			Children:  children,
			StartRow:  node.StartRow,
			StartCol:  node.StartCol,
			EndRow:    node.EndRow,
			EndCol:    node.EndCol,
			LineBreak: node.LineBreak,
			Space:     node.Space,
		}, true
	}

	newChildren := make([]NvimGuiNode, 0, len(node.Children))
	cursorInserted := false
	for _, child := range node.Children {
		if cursorInserted {
			// After cursor insertion, just copy remaining children
			newChildren = append(newChildren, child)
			continue
		}

		processedChild, inserted := processNode(child, cursorRow, cursorCol)
		newChildren = append(newChildren, processedChild)
		cursorInserted = inserted
	}

	return NvimGuiNode{
		Id:        node.Id,
		Text:      node.Text,
		HlGroup:   node.HlGroup,
		StartRow:  node.StartRow,
		StartCol:  node.StartCol,
		EndRow:    node.EndRow,
		EndCol:    node.EndCol,
		Root:      node.Root,
		LineBreak: node.LineBreak,
		Space:     node.Space,
		Children:  newChildren,
	}, cursorInserted
}

func splitTextAtCursor(node NvimGuiNode, cursorRow uint64, cursorCol uint64) (before, cursor, after NvimGuiNode) {
	if node.LineBreak {
		return handleLineBreakCursor(node, cursorRow, cursorCol)
	}

	if node.StartRow == node.EndRow {
		return handleTextNodeCursor(node, cursorRow, cursorCol)
	}
	return
}

func handleLineBreakCursor(node NvimGuiNode, cursorRow, cursorCol uint64) (before, cursor, after NvimGuiNode) {
	// Only create cursor node for line breaks
	cursor = NvimGuiNode{
		Id:        node.Id + "_cursor",
		Text:      "\n",
		HlGroup:   "cursor",
		StartRow:  cursorRow,
		StartCol:  cursorCol,
		EndRow:    cursorRow + 1,
		EndCol:    0,
		LineBreak: true,
	}
	return NvimGuiNode{}, cursor, NvimGuiNode{}
}

func handleTextNodeCursor(node NvimGuiNode, cursorRow, cursorCol uint64) (before, cursor, after NvimGuiNode) {
	relativeCol := cursorCol - node.StartCol
	runes := []rune(node.Text)

	// Handle edge cases
	if len(runes) == 0 || int(relativeCol) > len(runes) {
		return NvimGuiNode{}, NvimGuiNode{}, NvimGuiNode{}
	}

	// Create before node only if needed
	if relativeCol > 0 {
		before = createBeforeNode(node, relativeCol, cursorCol)
	}

	// Always create cursor node
	cursor = createCursorNode(node, relativeCol, cursorRow, cursorCol)

	// Create after node only if needed
	if int(relativeCol) < len(runes)-1 {
		after = createAfterNode(node, relativeCol, cursorCol)
	}

	return before, cursor, after
}

func createBeforeNode(node NvimGuiNode, relativeCol, cursorCol uint64) NvimGuiNode {
	return NvimGuiNode{
		Id:        node.Id + "_before",
		Text:      string([]rune(node.Text)[:relativeCol]),
		HlGroup:   node.HlGroup,
		StartRow:  node.StartRow,
		StartCol:  node.StartCol,
		EndRow:    node.StartRow,
		EndCol:    cursorCol,
		LineBreak: false,
		Space:     node.Space,
	}
}

func createCursorNode(node NvimGuiNode, relativeCol, cursorRow, cursorCol uint64) NvimGuiNode {
	return NvimGuiNode{
		Id:        node.Id + "_cursor",
		Text:      string([]rune(node.Text)[relativeCol : relativeCol+1]),
		HlGroup:   "cursor",
		StartRow:  cursorRow,
		StartCol:  cursorCol,
		EndRow:    cursorRow,
		EndCol:    cursorCol + 1,
		LineBreak: false,
		Space:     false,
	}
}

func createAfterNode(node NvimGuiNode, relativeCol, cursorCol uint64) NvimGuiNode {
	return NvimGuiNode{
		Id:        node.Id + "_after",
		Text:      string([]rune(node.Text)[relativeCol+1:]),
		HlGroup:   node.HlGroup,
		StartRow:  node.StartRow,
		StartCol:  cursorCol + 1,
		EndRow:    node.EndRow,
		EndCol:    node.EndCol,
		LineBreak: node.LineBreak,
		Space:     node.Space,
	}
}

func ApplyVisualSelection(root NvimGuiNode, selection NvimRange) NvimGuiNode {
	return processNodeForSelection(root, selection)
}

func isNodeInSelection(nodeRange, selection NvimRange) string {
	// Node completely before or after selection
	if (nodeRange.EndRow < selection.StartRow) ||
		(nodeRange.StartRow > selection.EndRow) ||
		(nodeRange.EndRow == selection.StartRow && nodeRange.EndCol <= selection.StartCol) ||
		(nodeRange.StartRow == selection.EndRow && nodeRange.StartCol >= selection.EndCol) {
		return "outside"
	}

	// Node completely inside selection
	if (nodeRange.StartRow > selection.StartRow ||
		(nodeRange.StartRow == selection.StartRow && nodeRange.StartCol >= selection.StartCol)) &&
		(nodeRange.EndRow < selection.EndRow ||
			(nodeRange.EndRow == selection.EndRow && nodeRange.EndCol <= selection.EndCol)) {
		return "inside"
	}

	// Node has some overlap with selection - this is the partial case
	return "partial"
}

func processNodeForSelection(node NvimGuiNode, selection NvimRange) NvimGuiNode {
	// Skip processing for nodes that already have cursor highlight
	if node.HlGroup == "cursor" {
		return node
	}

	nodeRange := NvimRange{
		StartRow: node.StartRow,
		StartCol: node.StartCol,
		EndRow:   node.EndRow,
		EndCol:   node.EndCol,
	}
	relation := isNodeInSelection(nodeRange, selection)

	switch relation {
	case "outside":
		// Node is completely outside selection, process children still
		if len(node.Children) > 0 {
			newChildren := make([]NvimGuiNode, len(node.Children))
			for i, child := range node.Children {
				newChildren[i] = processNodeForSelection(child, selection)
			}
			newNode := node
			newNode.Children = newChildren
			return newNode
		}
		// Leaf node outside selection, keep it unchanged
		return node
		
	case "inside":
		// Node is fully inside selection, update its highlight group
		newNode := node
		newNode.HlGroup = "visual-selection"
		
		// For fully-inside nodes with children, don't process children further
		// as the entire subtree will get the visual-selection highlight
		return newNode
		
	case "partial":
		// Node partially overlaps with selection
		if len(node.Children) > 0 {
			// Process children for partial nodes with children
			newChildren := make([]NvimGuiNode, len(node.Children))
			for i, child := range node.Children {
				newChildren[i] = processNodeForSelection(child, selection)
			}
			newNode := node
			newNode.Children = newChildren
			return newNode
		} else {
			// Split leaf node at selection boundaries
			return splitNodeForSelection(node, selection)
		}
	}
	
	return node
}

func splitSingleLineNode(node NvimGuiNode, selection NvimRange) []NvimGuiNode {
	// For nodes within a single row
	row := node.StartRow
	startCol := node.StartCol
	endCol := node.EndCol
	
	// Check if selection has any impact on this row
	if row < selection.StartRow || row >= selection.EndRow {
		return []NvimGuiNode{node}
	}
	
	// Determine the effective selection boundaries for this row
	var selStartCol, selEndCol uint64
	
	if row == selection.StartRow {
		selStartCol = selection.StartCol
	} else {
		// For rows after the first row of selection, start from beginning
		selStartCol = 0
	}
	
	if row == selection.EndRow-1 {
		selEndCol = selection.EndCol
	} else {
		// For rows before the last row of selection, go to end of row
		selEndCol = ^uint64(0) // Maximum value
	}
	
	// Calculate the effective boundaries of text we need to split
	effectiveStart := max(startCol, selStartCol)
	effectiveEnd := min(endCol, selEndCol)
	
	// If no effective overlap, return original node
	if effectiveStart >= effectiveEnd {
		return []NvimGuiNode{node}
	}
	
	// Split the text into three parts: before selection, inside selection, after selection
	runes := []rune(node.Text)
	parts := make([]NvimGuiNode, 0, 3)
	
	// 1. Text before selection
	if startCol < effectiveStart {
		beforeLen := effectiveStart - startCol
		if beforeLen > 0 && int(beforeLen) <= len(runes) {
			beforePart := NvimGuiNode{
				Id:        node.Id + "_before_sel",
				Text:      string(runes[:beforeLen]),
				HlGroup:   node.HlGroup, // Keep original highlight
				StartRow:  row,
				StartCol:  startCol,
				EndRow:    row,
				EndCol:    effectiveStart,
				LineBreak: false,
				Space:     node.Space,
			}
			parts = append(parts, beforePart)
		}
	}
	
	// 2. Text inside selection
	selectedLen := effectiveEnd - effectiveStart
	if selectedLen > 0 {
		startIdx := effectiveStart - startCol
		endIdx := effectiveEnd - startCol
		
		// Ensure valid indices
		if startIdx >= 0 && int(endIdx) <= len(runes) && startIdx < endIdx {
			selectedPart := NvimGuiNode{
				Id:        node.Id + "_sel",
				Text:      string(runes[startIdx:endIdx]),
				HlGroup:   "visual-selection", // Apply selection highlight
				StartRow:  row,
				StartCol:  effectiveStart,
				EndRow:    row,
				EndCol:    effectiveEnd,
				LineBreak: false,
				Space:     node.Space,
			}
			parts = append(parts, selectedPart)
		}
	}
	
	// 3. Text after selection
	if effectiveEnd < endCol {
		afterIdx := effectiveEnd - startCol
		if afterIdx >= 0 && int(afterIdx) < len(runes) {
			afterPart := NvimGuiNode{
				Id:        node.Id + "_after_sel",
				Text:      string(runes[afterIdx:]),
				HlGroup:   node.HlGroup, // Keep original highlight
				StartRow:  row,
				StartCol:  effectiveEnd,
				EndRow:    row,
				EndCol:    endCol,
				LineBreak: false,
				Space:     node.Space,
			}
			parts = append(parts, afterPart)
		}
	}
	
	// If we've created multiple parts, wrap them in a container
	if len(parts) > 1 {
		return parts
	} else if len(parts) == 1 {
		// If we ended up with just one part, return it directly
		return parts
	}
	
	// Fallback to original node if something went wrong
	return []NvimGuiNode{node}
}

func handleLineBreakSelection(node NvimGuiNode, selection NvimRange) NvimGuiNode {
	// Line breaks are special - they're either entirely in or out of selection
	// Based on node's row (start row) position relative to selection

	if node.StartRow < selection.StartRow || node.StartRow >= selection.EndRow {
		// Line break outside selection
		return node
	}

	// Line break inside selection
	newNode := node
	newNode.HlGroup = "visual-selection"
	return newNode
}

func splitNodeForSelection(node NvimGuiNode, selection NvimRange) NvimGuiNode {
	// Handle line breaks specially
	if node.LineBreak {
		return handleLineBreakSelection(node, selection)
	}
	
	// For text nodes, determine which parts are in selection
	var parts []NvimGuiNode
	
	if node.StartRow == node.EndRow {
		// Single-line node processing
		parts = splitSingleLineNode(node, selection)
	} else {
		// Multi-line nodes might need special handling
		// For now, we might want to return the node as is or process it line by line
		// This would be a more complex implementation
		parts = []NvimGuiNode{node}
	}
	
	// If we only got one part or no parts, return original or that part
	if len(parts) <= 0 {
		return node
	}
	if len(parts) == 1 {
		return parts[0]
	}
	
	// Create a container with all the split parts
	return NvimGuiNode{
		Id:        node.Id + "_sel_container",
		HlGroup:   node.HlGroup,
		StartRow:  node.StartRow,
		StartCol:  node.StartCol,
		EndRow:    node.EndRow,
		EndCol:    node.EndCol,
		LineBreak: node.LineBreak,
		Space:     node.Space,
		Children:  parts,
	}
}

// Helper functions
func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
