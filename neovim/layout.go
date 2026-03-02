package neovim

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
	dirty          bool
}

func NewGridLayout() *GridLayout {
	return &GridLayout{
		Cols:           "1fr",
		Rows:           "1fr",
		ActiveWindowId: 1000,
		Windows:        []*WindowAPI{},
		dirty:          true,
	}
}

func (s *Screen) CalculateGridLayout() {
	// Step 1: Collect all unique row and column positions
	rowPositions := make(map[int]bool)
	colPositions := make(map[int]bool)

	for _, window := range s.Windows {
		if window.IsFloating() {
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
	totalHeight := s.Height - 1
	// Step 3: Calculate the CSS grid template strings
	colsFractions := calculateFractions(cols, s.Width)
	rowsFractions := calculateFractions(rows, totalHeight)
	colsStr := strings.Join(utils.MapArray(colsFractions, func(i int) string { return strconv.Itoa(i) }), "fr ") + "fr"
	rowsStr := strings.Join(utils.MapArray(rowsFractions, func(i int) string { return strconv.Itoa(i) }), "fr ") + "fr"

	// Step 4: Calculate the grid position for each window
	windowAPIs := make([]*WindowAPI, 0, len(s.Windows))

	var windows []*Window
	var floatingWindows []*Window
	for winId, window := range s.Windows {
		if window.Hidden || window.IsFloating() {
			floatingWindows = append(floatingWindows, window)
			continue
		}
		windows = append(windows, window)

		// Find grid indices for this window
		// colStart := findIndex(cols, window.StartCol)
		// window.StartCol == 0 -> colStart 1

		// window.EndCol == s.Width -> colStart len(colsFractions)+1
		colStart := findStartIndex(colsFractions, window.StartCol)
		rowStart := findStartIndex(rowsFractions, window.StartRow)
		colEnd := findEndIndex(colsFractions, window.StartCol+window.Width, s.Width)
		rowEnd := findEndIndex(rowsFractions, window.StartRow+window.Height, totalHeight)

		width := window.Width
		height := window.Height

		w := &WindowAPI{
			ID:                  winId,
			Type:                window.Type,
			Width:               utils.CalculatePercentage(width, s.Width),
			Height:              utils.CalculatePercentage(height, s.Height),
			ColStart:            colStart,
			ColEnd:              colEnd,
			RowStart:            rowStart,
			RowEnd:              rowEnd,
			Filetype:            "",
			Filepath:            "",
			LineNumbers:         window.lineNumbers,
			RelativeLineNumbers: window.relativeLineNumbers,
			Mode:                window.Mode,
			Cursor:              window.Cursor,
		}
		if window.Buffer != nil {
			buf := *window.Buffer
			w.Filetype = buf.Filetype
			w.Filepath = buf.Filepath
		}
		w.ColorColumns = s.ColorColumns
		w.ColorColumnColor = s.ColorColumnColor
		if s.CursorLineEnabled {
			w.CursorLineColor = s.CursorLineColor
		}

		windowAPIs = append(windowAPIs, w)
	}
	sort.SliceStable(windowAPIs, func(i, j int) bool { return windowAPIs[i].ID < windowAPIs[j].ID })

	windowAPIsChanged := false
	if len(s.layout.Windows) == len(windowAPIs) {
		for i, windowAPI := range windowAPIs {
			if s.layout.Windows[i].Filepath != windowAPI.Filepath {
				windowAPIsChanged = true
				break
			}
		}
	} else {
		windowAPIsChanged = true
	}
	if s.layout.Cols != colsStr || s.layout.Rows != rowsStr || windowAPIsChanged {
		s.layout.dirty = true
	}

	s.layout.Cols = colsStr
	s.layout.Rows = rowsStr
	s.layout.Windows = windowAPIs
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
