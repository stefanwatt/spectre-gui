package rendering

import (
	"nvim-gui/utils"
	"sort"
	"strconv"
	"strings"
)

type GridLayout struct {
	Cols           string       `json:"cols"`
	Rows           string       `json:"rows"`
	ActiveWindowId int          `json:"activeWindowId"`
	Windows        []*WindowAPI `json:"windows"`
}

//TODO: check if this isnt just a stupid duplication of windowapi
type LayoutWindow struct {
	ID                  int
	Type                string
	Hidden              bool
	Floating            bool
	StartRow            int
	StartCol            int
	Width               int
	Height              int
	LineNumbers         bool
	RelativeLineNumbers bool
	Mode                string
	Cursor              *Cursor
	Filetype            string
	Filepath            string
}

type LayoutInput struct {
	ScreenWidth       int
	ScreenHeight      int
	ActiveWindowID    int
	ColorColumns      []int
	ColorColumnColor  string
	CursorLineEnabled bool
	CursorLineColor   string
	Windows           []LayoutWindow
}

func NewGridLayout() *GridLayout {
	return &GridLayout{
		Cols:           "1fr",
		Rows:           "1fr",
		ActiveWindowId: 1000,
		Windows:        []*WindowAPI{},
	}
}

func CalculateGridLayout(input LayoutInput) *GridLayout {
	// Step 1: Collect all unique row and column positions
	rowPositions := make(map[int]bool)
	colPositions := make(map[int]bool)

	for _, window := range input.Windows {
		if window.Floating || window.Hidden {
			continue
		}
		rowPositions[window.StartRow] = true
		colPositions[window.StartCol] = true
	}

	// Step 2: Convert to sorted slices
	rows := make([]int, 0, len(rowPositions))
	cols := make([]int, 0, len(colPositions))

	for pos := range rowPositions {
		rows = append(rows, pos)
	}
	for pos := range colPositions {
		cols = append(cols, pos)
	}

	sort.Ints(rows)
	sort.Ints(cols)

	// subtract statusline
	totalHeight := input.ScreenHeight - 1
	if totalHeight < 1 {
		totalHeight = 1
	}
	// Step 3: Calculate the CSS grid template strings
	colsFractions := calculateFractions(cols, input.ScreenWidth)
	rowsFractions := calculateFractions(rows, totalHeight)
	colsStr := strings.Join(utils.MapArray(colsFractions, func(i int) string { return strconv.Itoa(i) }), "fr ") + "fr"
	rowsStr := strings.Join(utils.MapArray(rowsFractions, func(i int) string { return strconv.Itoa(i) }), "fr ") + "fr"

	// Step 4: Calculate the grid position for each window
	windowAPIs := make([]*WindowAPI, 0, len(input.Windows))

	for _, window := range input.Windows {
		if window.Hidden || window.Floating {
			continue
		}
		winID := window.ID

		// Find grid indices for this window
		// colStart := findIndex(cols, window.StartCol)
		// window.StartCol == 0 -> colStart 1

		// window.EndCol == s.Width -> colStart len(colsFractions)+1
		colStart := findStartIndex(colsFractions, window.StartCol)
		rowStart := findStartIndex(rowsFractions, window.StartRow)
		colEnd := findEndIndex(colsFractions, window.StartCol+window.Width, input.ScreenWidth)
		rowEnd := findEndIndex(rowsFractions, window.StartRow+window.Height, totalHeight)

		width := window.Width
		height := window.Height

		w := &WindowAPI{
			ID:                  winID,
			Type:                window.Type,
			Width:               utils.CalculatePercentage(width, input.ScreenWidth),
			Height:              utils.CalculatePercentage(height, input.ScreenHeight),
			ColStart:            colStart,
			ColEnd:              colEnd,
			RowStart:            rowStart,
			RowEnd:              rowEnd,
			Filetype:            window.Filetype,
			Filepath:            window.Filepath,
			LineNumbers:         window.LineNumbers,
			RelativeLineNumbers: window.RelativeLineNumbers,
			Mode:                window.Mode,
			Cursor:              window.Cursor,
		}
		w.ColorColumns = input.ColorColumns
		w.ColorColumnColor = input.ColorColumnColor
		if input.CursorLineEnabled {
			w.CursorLineColor = input.CursorLineColor
		}

		windowAPIs = append(windowAPIs, w)
	}
	sort.SliceStable(windowAPIs, func(i, j int) bool { return windowAPIs[i].ID < windowAPIs[j].ID })

	return &GridLayout{
		Cols:           colsStr,
		Rows:           rowsStr,
		ActiveWindowId: input.ActiveWindowID,
		Windows:        windowAPIs,
	}
}

func LayoutEqual(a, b *GridLayout) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Cols != b.Cols || a.Rows != b.Rows || a.ActiveWindowId != b.ActiveWindowId {
		return false
	}
	if len(a.Windows) != len(b.Windows) {
		return false
	}
	for i := range a.Windows {
		if !a.Windows[i].Equal(b.Windows[i]) {
			return false
		}
		if len(a.Windows[i].ColorColumns) != len(b.Windows[i].ColorColumns) {
			return false
		}
		for j := range a.Windows[i].ColorColumns {
			if a.Windows[i].ColorColumns[j] != b.Windows[i].ColorColumns[j] {
				return false
			}
		}
	}
	return true
}

// calculateFractions constructs CSS grid-template string
func calculateFractions(positions []int, totalSize int) []int {
	if len(positions) <= 1 {
		return []int{1}
	}
	result := []int{}
	for i := 0; i < len(positions)-1; i++ {
		size := positions[i+1] - positions[i]
		result = append(result, size)
	}
	result = append(result, totalSize-positions[len(positions)-1])

	return result
}

// findStartIndex finds the index of a value in a sorted slice
// findStartIndex finds the CSS grid line for the window's start position (inclusive)
func findStartIndex(fractions []int, val int) int {
	cumulative := 0
	for line, size := range fractions {
		if val <= cumulative {
			return line + 1 // CSS lines start at 1
		}
		cumulative += size
	}
	return len(fractions) + 1
}

// findEndIndex finds the CSS grid line for the window's end position (exclusive)
func findEndIndex(fractions []int, val int, max int) int {
	if val >= max {
		return len(fractions) + 1
	}
	cumulative := 0
	for line, size := range fractions {
		cumulative += size
		if val < cumulative {
			return line + 2 // End line is next after the track containing val
		}
	}
	return len(fractions) + 1
}
