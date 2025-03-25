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

type CursorState struct {
	Row uint64
	Col uint64
}

type SelectionState struct {
	Active bool
	Range  NvimRange
}

var (
	currentTree    NvimGuiNode
	cursorState    CursorState
	selectionState SelectionState
	currentMode    string
	NvimInstance   *nvim.Nvim
)

func StartListening(servername string, ctx context.Context) {
	var err error
	NvimInstance, err = nvim.NewChildProcess(
		nvim.ChildProcessCommand("nvim"),
		nvim.ChildProcessArgs("--embed", "/tmp/foo.lua"),
		nvim.ChildProcessContext(context.Background()),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer NvimInstance.Close()
	var result string

	opts := make(map[string]interface{})
	opts["ext_cmdline"] = true
	err = NvimInstance.AttachUI(100, 40, opts)
	if err != nil {
		utils.Log(err.Error())
	}
	NvimInstance.RegisterHandler("redraw", func(updates ...[]interface{}) {
		for _, update := range updates {
			if len(update) == 0 {
				continue
			}

			event, ok := update[0].(string)
			if !ok {
				continue
			}

			if event == "cmdline_show" {
				Runtime.EventsEmit(ctx, "cmdline_show")
			}
			if event == "cmdline_hide" {
				Runtime.EventsEmit(ctx, "cmdline_hide")
			}
		}
	})
	nvim_cmd := fmt.Sprintf("return require('nvim-gui.foo').attach_buffer(%d)", NvimInstance.ChannelID())
	err = NvimInstance.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}

	nvim_cmd = fmt.Sprintf("return require('nvim-gui.foo').listen_for_visual_selection_change(%d)", NvimInstance.ChannelID())
	err = NvimInstance.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	nvim_cmd = fmt.Sprintf("return require('nvim-gui.foo').listen_for_cursor_move(%d)", NvimInstance.ChannelID())
	err = NvimInstance.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	nvim_cmd = fmt.Sprintf("return require('nvim-gui.foo').listen_for_mode_change(%d)", NvimInstance.ChannelID())
	err = NvimInstance.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}

	// Initialize default state
	cursorState = CursorState{0, 0}
	selectionState = SelectionState{Active: false, Range: NvimRange{}}
	currentMode = "n" // Default to normal mode

	NvimInstance.RegisterHandler("nvim-gui-buf-changed", func(v *nvim.Nvim, root NvimGuiNode) {
		// Store the new tree
		currentTree = root

		// Process with both selection and cursor
		processedTree := processTreeWithSelectionAndCursor(currentTree)

		// Send the fully processed tree
		OnBufChanged(ctx, processedTree)
	})

	// Update the nvim-gui-visual-selection-changed handler
	NvimInstance.RegisterHandler("nvim-gui-visual-selection-changed", func(v *nvim.Nvim, selection_range NvimRange) {
		// Normalize selection start and end
		startRow, startCol, endRow, endCol := normalizeSelection(selection_range)
		normalizedSelection := NvimRange{
			StartRow: startRow,
			StartCol: startCol,
			EndRow:   endRow,
			EndCol:   endCol,
		}

		// Update selection state
		selectionState.Range = normalizedSelection
		selectionState.Active = true

		// Process the tree with updated state
		processedTree := processTreeWithSelectionAndCursor(currentTree)

		// Only process if we're in visual mode
		if isVisualMode(currentMode) {
			// Send the processed tree
			OnBufChanged(ctx, processedTree)
		}

		// Notify about selection change
		Runtime.EventsEmit(ctx, "visual-selection-changed", normalizedSelection)

		utils.Log(fmt.Sprintf("Visual selection changed: (%d:%d) -> (%d:%d)",
			startRow, startCol, endRow, endCol))
	})

	NvimInstance.RegisterHandler("nvim-gui-cursor-moved", func(v *nvim.Nvim, cursor_move_event CursorMoveEvent) {
		// Update cursor state
		cursorState.Row = cursor_move_event.Row
		cursorState.Col = cursor_move_event.Col

		// In visual mode, cursor movement also updates selection end point
		if isVisualMode(currentMode) && selectionState.Active {
			selectionState.Range.EndRow = cursorState.Row
			selectionState.Range.EndCol = cursorState.Col

			// Normalize the selection range
			startRow, startCol, endRow, endCol := normalizeSelection(selectionState.Range)
			selectionState.Range = NvimRange{
				StartRow: startRow,
				StartCol: startCol,
				EndRow:   endRow,
				EndCol:   endCol,
			}
		}

		// Process the tree with updated state
		processedTree := processTreeWithSelectionAndCursor(currentTree)

		// Debug output
		utils.Log(fmt.Sprintf("Cursor moved to %d:%d in mode %s",
			cursorState.Row, cursorState.Col, currentMode))

		// Send updates
		OnBufChanged(ctx, processedTree)
		UpdateCursor(ctx, cursor_move_event)
	})

	NvimInstance.RegisterHandler("nvim-gui-mode-changed", func(v *nvim.Nvim, args []string) {
		mode := args[0]
		currentMode = mode

		// If exiting visual mode, clear the selection state
		if !isVisualMode(mode) {
			selectionState.Active = false
		}

		// Process the tree with updated state
		processedTree := processTreeWithSelectionAndCursor(currentTree)

		// Update the mode and refresh the UI
		UpdateSelection(ctx, mode)
		OnBufChanged(ctx, processedTree)

		utils.Log(fmt.Sprintf("Mode changed to: %s", mode))
	})

	if err := NvimInstance.Serve(); err != nil {
		log.Fatal(err)
	}
	log.Println("listening terminating")
}

// Helper function to check if a mode is visual mode
func isVisualMode(mode string) bool {
	return mode == "v" || mode == "V" || mode == "\x16" // Normal, line, and block visual modes
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

type Position struct {
	Row, Col uint64
}

func (p Position) Less(other Position) bool {
	return p.Row < other.Row || (p.Row == other.Row && p.Col < other.Col)
}

func (p Position) Greater(other Position) bool {
	return p.Row > other.Row || (p.Row == other.Row && p.Col > other.Col)
}

func (p Position) LessOrEqual(other Position) bool {
	return p.Less(other) || p == other
}

func (p Position) GreaterOrEqual(other Position) bool {
	return p.Greater(other) || p == other
}

func normalizeSelection(selection NvimRange) (uint64, uint64, uint64, uint64) {
	start := Position{selection.StartRow, selection.StartCol}
	end := Position{selection.EndRow, selection.EndCol}
	if start.Greater(end) {
		return end.Row, end.Col, start.Row, start.Col
	}
	return start.Row, start.Col, end.Row, end.Col
}

func processTreeWithSelectionAndCursor(tree NvimGuiNode) NvimGuiNode {
	processedTree := tree

	// Step 1: Apply visual selection if applicable
	if isVisualMode(currentMode) && selectionState.Active {
		processedTree = ApplyVisualSelection(processedTree, selectionState.Range)
	}

	// Step 2: Apply cursor (always done after selection)
	processedTree = SplitNodeAtCursor(processedTree, cursorState.Row, cursorState.Col)

	return processedTree
}
