package fileexplorer

import (
	"nvim-gui/core/model"
	"strings"
)

// DirectoryEntry is a parsed file/directory entry from a mini.files pane.
type DirectoryEntry struct {
	ID        int    `json:"id"`
	Icon      string `json:"icon"`
	IconClass string `json:"iconClass"`
	Text      string `json:"text"`
	IsDir     bool   `json:"isDir"`
}

// Directory holds parsed entries for one mini.files pane (parent/current/preview).
type Directory struct {
	WinID           int              `json:"winId"`
	BufNr           int              `json:"bufNr"`
	Entries         []DirectoryEntry `json:"entries"`
	SelectedEntryId int              `json:"selectedEntryId"`
	CursorCol       int              `json:"cursorCol"`
}

// ParseDirectoryEntries parses a grid's cell data into directory entries.
// cells is the grid's [][]*model.Cell, height is the number of rows to process.
// directoryLineMap maps (1-indexed) line numbers to whether they are directories;
// if a line is not present in the map we fall back to checking for a trailing "/".
func ParseDirectoryEntries(cells [][]*model.Cell, height int, directoryLineMap map[int]bool) []DirectoryEntry {
	entries := make([]DirectoryEntry, 0, height)
	for row := 0; row < height; row++ {
		if row >= len(cells) {
			continue
		}
		rowCells := cells[row]
		if len(rowCells) == 0 {
			continue
		}

		// Find first and last non-empty cells
		firstNonEmpty := -1
		lastNonEmpty := -1
		for i, cell := range rowCells {
			if cell != nil && strings.TrimSpace(cell.Char) != "" {
				firstNonEmpty = i
				break
			}
		}
		if firstNonEmpty == -1 {
			continue
		}
		for i := len(rowCells) - 1; i >= 0; i-- {
			if rowCells[i] != nil && strings.TrimSpace(rowCells[i].Char) != "" {
				lastNonEmpty = i
				break
			}
		}
		if lastNonEmpty == -1 {
			continue
		}

		leftChar := rowCells[firstNonEmpty].Char
		rightChar := rowCells[lastNonEmpty].Char

		// Skip border rows (┌─── or └───)
		if leftChar == "┌" || leftChar == "└" {
			continue
		}

		// Strip │ border characters if present
		contentStart := firstNonEmpty
		contentEnd := lastNonEmpty
		if leftChar == "│" && rightChar == "│" && contentEnd-contentStart >= 2 {
			contentStart++
			contentEnd--
		}
		if contentStart > contentEnd {
			continue
		}

		contentCells := rowCells[contentStart : contentEnd+1]
		line := strings.TrimSpace(cellsToString(contentCells))
		if line == "" {
			continue
		}

		// Extract icon (first non-space cell that has a non-ASCII rune)
		var icon string
		var iconClass string
		text := line
		iconCellIndex := -1
		for i, cell := range contentCells {
			if cell != nil && strings.TrimSpace(cell.Char) != "" {
				iconCellIndex = i
				break
			}
		}
		if iconCellIndex >= 0 {
			iconRunes := []rune(contentCells[iconCellIndex].Char)
			if len(iconRunes) > 0 && iconRunes[0] > 127 {
				icon = contentCells[iconCellIndex].Char
				iconClass = contentCells[iconCellIndex].ClassesToString()
				text = strings.TrimSpace(cellsToString(contentCells[iconCellIndex+1:]))
			}
		}

		if icon == "" && text == "" {
			continue
		}

		// Determine if entry is a directory
		isDir, ok := directoryLineMap[row]
		if !ok {
			isDir = strings.HasSuffix(text, "/")
		}

		entries = append(entries, DirectoryEntry{
			ID:        row,
			Icon:      icon,
			IconClass: iconClass,
			Text:      text,
			IsDir:     isDir,
		})
	}
	return entries
}

// GridToString converts a grid's cells into a single string with newlines between rows.
// Used by confirm prompt detection.
func GridToString(cells [][]*model.Cell) string {
	var builder strings.Builder
	for _, row := range cells {
		for _, cell := range row {
			if cell == nil {
				continue
			}
			builder.WriteString(cell.Char)
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

// DetectConfirmPrompt checks if a floating window's grid content looks like
// the mini.files "close without synchronization" confirm dialog.
// Returns the prompt info if detected.
type ConfirmPrompt struct {
	WinID   int      `json:"winId"`
	Message string   `json:"message"`
	Choices []string `json:"choices"`
}

func DetectConfirmPrompt(winID int, cells [][]*model.Cell, filetype string) (*ConfirmPrompt, bool) {
	if filetype != "" {
		return nil, false
	}
	raw := strings.TrimSpace(GridToString(cells))
	if raw == "" || !strings.Contains(raw, "without synchronization") || !strings.Contains(raw, "Confirm") {
		return nil, false
	}

	lines := strings.Split(raw, "\n")
	cleanLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "│") || strings.Contains(trimmed, "─") || strings.Contains(trimmed, "┌") || strings.Contains(trimmed, "└") {
			trimmed = strings.Map(func(r rune) rune {
				switch r {
				case '│', '─', '┌', '┐', '└', '┘':
					return -1
				default:
					return r
				}
			}, trimmed)
			trimmed = strings.TrimSpace(trimmed)
		}
		if trimmed == "" {
			continue
		}
		trimmed = strings.ReplaceAll(trimmed, "&", "")
		if trimmed == "Yes" || trimmed == "No" || strings.HasPrefix(trimmed, "Yes") || strings.HasPrefix(trimmed, "No") {
			continue
		}
		cleanLines = append(cleanLines, trimmed)
	}

	message := strings.Join(cleanLines, "\n")
	if strings.TrimSpace(message) == "" {
		message = "Confirm close without synchronization?"
	}

	return &ConfirmPrompt{
		WinID:   winID,
		Message: message,
		Choices: []string{"Yes", "No"},
	}, true
}

// cellsToString concatenates cell chars into a string.
func cellsToString(cells []*model.Cell) string {
	var builder strings.Builder
	for _, cell := range cells {
		if cell == nil {
			continue
		}
		builder.WriteString(cell.Char)
	}
	return builder.String()
}
