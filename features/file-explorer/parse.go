package fileexplorer

import (
	"nvim-gui/core/model"
	"regexp"
	"strings"
)

// borderChars contains all box-drawing characters used by floating window borders.
const borderChars = "│┌┐└┘─"

// entryRegex matches a directory entry line: an optional non-ASCII icon followed
// by whitespace and the filename. The icon is a single multi-byte character with
// codepoint > 127 that is NOT a box-drawing border character.
var entryRegex = regexp.MustCompile(`^(\S+)\s+(.+)$`)

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

		// Convert the entire row to a string and strip border characters.
		// This is robust against cursor-row highlight changes that can break
		// cell-level left/right border detection.
		rawLine := cellsToString(rowCells)
		line := stripBorderChars(rawLine)
		if line == "" {
			continue
		}

		// Use regex to extract icon and filename from the content.
		// Entry format: "<icon> <filename>" where icon is a non-ASCII character.
		var icon string
		var iconClass string
		var text string

		match := entryRegex.FindStringSubmatch(line)
		if match != nil {
			candidate := match[1]
			candidateRunes := []rune(candidate)
			// Icon must be a single non-ASCII character that is not a border char
			if len(candidateRunes) == 1 && candidateRunes[0] > 127 && !strings.ContainsRune(borderChars, candidateRunes[0]) {
				icon = candidate
				text = strings.TrimSpace(match[2])
				// Walk the cells to find the icon cell and extract its CSS class
				iconClass = findCellClass(rowCells, icon)
			} else {
				// First token is not an icon — treat the whole line as text
				text = line
			}
		} else {
			text = line
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

// stripBorderChars removes box-drawing border characters and surrounding
// whitespace from a row string. This handles all border patterns (┌─┐, └─┘,
// │...│) regardless of cursor highlighting or cell layout.
func stripBorderChars(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ""
	}
	// Skip entire border rows (top/bottom borders)
	firstRune := []rune(trimmed)[0]
	if firstRune == '┌' || firstRune == '└' || firstRune == '─' {
		return ""
	}
	// Strip leading and trailing │ (side borders)
	trimmed = strings.TrimLeft(trimmed, "│")
	trimmed = strings.TrimRight(trimmed, "│")
	return strings.TrimSpace(trimmed)
}

// findCellClass scans a row of cells for the first cell whose Char matches
// the target string and returns its CSS classes.
func findCellClass(cells []*model.Cell, target string) string {
	for _, cell := range cells {
		if cell != nil && cell.Char == target {
			return cell.ClassesToString()
		}
	}
	return ""
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
