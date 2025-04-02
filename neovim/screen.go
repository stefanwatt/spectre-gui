package neovim

import (
	"context"
	"fmt"
	"math"
	"nvim-gui/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akiyosi/goneovim/util"
	"github.com/neovim/go-client/nvim"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Screen struct {
	margins       []int
	Height        int //in number of cells
	Width         int //in number of cells
	topLine       int
	botLine       int
	curLine       int
	lineCount     int
	scrollDelta   float64
	ctx           context.Context
	Grids         map[int]*Grid
	Highlights    map[int]*Highlight
	Windows       map[int]*Window // Map of window IDs to Window objects
	GridToWindow  map[int]int     // Map of grid IDs to window IDs
	DefaultFg     int
	DefaultBg     int
	DefaultSp     int
	ActiveGrid    int
	ActiveWindow  int
	Mode          string
	PendingRender bool
	highlightsMu  sync.RWMutex // Mutex for Highlights map
	windowsMu     sync.RWMutex // Mutex for Windows map
}

func NewScreen(ctx context.Context, cols int, rows int) *Screen {
	content := make([][]*Cell, rows)
	for i := range content {
		content[i] = make([]*Cell, cols)
		for j := range content[i] {
			content[i][j] = &Cell{
				Char:      " ",
				Highlight: 0,
			}
		}
	}

	// Create default grid
	defaultGrid := &Grid{
		ID:     1,
		Width:  cols,
		Height: rows,
		Cells:  content,
	}

	// Create grids map and add default grid
	grids := make(map[int]*Grid)
	grids[1] = defaultGrid

	// Create default highlight
	highlights := make(map[int]*Highlight)
	highlights[0] = &Highlight{
		Foreground: 0xffffff,
		Background: 0x000000,
		Special:    0xffffff,
	}

	return &Screen{
		ctx:           ctx,
		Width:         cols,
		Height:        rows,
		Grids:         grids,
		Highlights:    highlights,
		Windows:       make(map[int]*Window),
		GridToWindow:  make(map[int]int),
		DefaultFg:     0xffffff,
		DefaultBg:     0x000000,
		DefaultSp:     0xffffff,
		ActiveGrid:    2,
		Mode:          "normal",
		PendingRender: false,
		margins:       make([]int, 4), // Initialize margins slice
	}
}

func handleEvent(update interface{}) (event string, ok bool) {
	switch update.(type) {
	case string:
		event = update.(string)
		ok = true
	default:
		event = ""
		ok = false
	}

	return event, ok
}

func (s *Screen) Resize(width int, height int) {
	screen.Width = width
	screen.Height = height
	NvimInstance.TryResizeUI(width, height)
}

func (s *Screen) handleRedraw(updates [][]interface{}) {
	for _, update := range updates {
		if len(update) == 0 {
			continue
		}
		event, ok := handleEvent(update[0])
		if !ok {
			continue
		}

		args := update[1:]

		switch event {
		case "flush":
			s.scheduleRender()
		case "grid_resize":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(gridArgs[0])
				width := utils.ReflectToInt(gridArgs[1])
				height := utils.ReflectToInt(gridArgs[2])
				s.gridResize(gridId, width, height)
			}
		case "grid_line":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(gridArgs[0])
				row := utils.ReflectToInt(gridArgs[1])
				col := utils.ReflectToInt(gridArgs[2])
				cells := gridArgs[3].([]interface{})
				s.gridLine(gridId, row, col, cells)
			}
		case "grid_clear":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(gridArgs[0])
				s.gridClear(gridId)
			}
		case "grid_scroll":
			for _, arg := range args {
				scrollArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(scrollArgs[0])
				top := utils.ReflectToInt(scrollArgs[1])
				bot := utils.ReflectToInt(scrollArgs[2])
				left := utils.ReflectToInt(scrollArgs[3])
				right := utils.ReflectToInt(scrollArgs[4])
				rows := utils.ReflectToInt(scrollArgs[5])
				cols := utils.ReflectToInt(scrollArgs[6])
				s.gridScroll(gridId, top, bot, left, right, rows, cols)
			}
		case "grid_cursor_goto":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				gridId := utils.ReflectToInt(gridArgs[0])
				row := utils.ReflectToInt(gridArgs[1])
				col := utils.ReflectToInt(gridArgs[2])
				s.gridCursorGoto(gridId, row, col)
			}
		case "win_viewport":
			s.handleWinViewport(args)
		case "win_viewport_margins":
			s.handleWinViewportMargins(args)
		case "default_colors_set":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				fg := utils.ReflectToInt(gridArgs[0])
				bg := utils.ReflectToInt(gridArgs[1])
				sp := utils.ReflectToInt(gridArgs[2])
				s.defaultColorsSet(fg, bg, sp)
			}
		case "hl_attr_define":
			s.hlAttrDefine(args)
		case "mode_change":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				mode, _ := gridArgs[0].(string)
				s.modeChange(mode)
			}
		case "win_pos":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				s.winPos(gridArgs)
			}
		case "win_float_pos":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				s.winFloatPos(gridArgs)
			}
		case "win_close":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				s.winClose(gridArgs)
			}
		case "cmdline_show":
			s.handleCmdlineShow(args)
		case "cmdline_pos":
			s.handleCmdlinePos(args)
		case "cmdline_hide":
			Runtime.EventsEmit(s.ctx, "cmdline_hide")
		}
	}
}

func (s *Screen) gridScroll(gridId int, top int, bot int, left int, right int, rows int, cols int) {

	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}

	// Handle vertical scrolling
	if rows != 0 {
		direction := 1
		if rows < 0 {
			direction = -1
		}
		absRows := int(math.Abs(float64(rows)))

		if direction < 0 {
			// Scroll down (content moves up)
			for row := bot - 1; row >= top+absRows; row-- {
				for col := left; col < right; col++ {
					if row < grid.Height && col < grid.Width && row-absRows >= 0 {
						grid.Cells[row][col] = grid.Cells[row-absRows][col]
					}
				}
			}
			// Clear the newly exposed area
			for row := top; row < top+absRows && row < grid.Height; row++ {
				for col := left; col < right && col < grid.Width; col++ {
					grid.Cells[row][col] = &Cell{
						Char:      " ",
						Highlight: 0,
					}
				}
			}
		} else {
			// Scroll up (content moves down)
			for row := top; row < bot-absRows; row++ {
				for col := left; col < right; col++ {
					if row < grid.Height && col < grid.Width && row+absRows < grid.Height {
						grid.Cells[row][col] = grid.Cells[row+absRows][col]
					}
				}
			}
			// Clear the newly exposed area
			for row := bot - absRows; row < bot && row < grid.Height; row++ {
				for col := left; col < right && col < grid.Width; col++ {
					grid.Cells[row][col] = &Cell{
						Char:      " ",
						Highlight: 0,
					}
				}
			}
		}
	}
}

func (s *Screen) gridLine(gridId int, row int, col int, cells []interface{}) {
	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}
	if row >= grid.Height || col >= grid.Width {
		utils.Log(fmt.Sprintf("Row %d or col %d out of bounds for grid %d (max: %d,%d)",
			row, col, gridId, grid.Height-1, grid.Width-1))
		return
	}
	currentCol := col
	lastHl := 0
	if len(cells) == 0 {
		return
	}

	winId, exists := s.GridToWindow[gridId]
	if exists {
		s.windowsMu.Lock()
		window, _ := s.Windows[winId]
		window.Dirty = true
		s.windowsMu.Unlock()
	}
	for _, cell := range cells {
		cellData, ok := cell.([]interface{})
		if !ok || len(cellData) == 0 {
			continue
		}

		char := cellData[0].(string)

		repeat := 1
		if len(cellData) > 2 {
			repeat = utils.ReflectToInt(cellData[2])
		}

		hl := 0
		if len(cellData) > 1 {
			hl = utils.ReflectToInt(cellData[1])
			lastHl = hl
		}
		for i := 0; i < repeat && currentCol < grid.Width; i++ {
			if row < len(grid.Cells) && currentCol < len(grid.Cells[row]) {
				if hl == 0 {
					hl = lastHl
				}
				newCell := Cell{
					Char:      char,
					Highlight: hl,
				}
				grid.Cells[row][currentCol] = &newCell
				currentCol++
			}
		}
	}
}

func (s *Screen) gridClear(gridId int) {
	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}
	for i := range grid.Cells {
		for j := range grid.Cells[i] {
			grid.Cells[i][j] = &Cell{
				Char:      " ",
				Highlight: 0,
			}
		}
	}
	// s.scheduleRender()
}

func (s *Screen) handleWinViewport(args []interface{}) {
	if len(args) < 1 {
		return
	}

	viewportArgs, ok := args[0].([]interface{})
	if !ok || len(viewportArgs) < 8 {
		return
	}

	grid := utils.ReflectToInt(viewportArgs[0])
	if grid != 1 {
		return // Only care about main grid
	}

	// win := utils.ReflectToInt(viewportArgs[1]) // Window ID
	s.topLine = utils.ReflectToInt(viewportArgs[2])
	s.botLine = utils.ReflectToInt(viewportArgs[3])
	s.curLine = utils.ReflectToInt(viewportArgs[4])
	// curcol := utils.ReflectToInt(viewportArgs[5])
	s.lineCount = utils.ReflectToInt(viewportArgs[6])

	// Convert to float64 if it's a float
	if f, ok := viewportArgs[7].(float64); ok {
		s.scrollDelta = f
	} else {
		s.scrollDelta = float64(utils.ReflectToInt(viewportArgs[7]))
	}

	// Emit viewport info to frontend
	Runtime.EventsEmit(s.ctx, "viewport_changed", map[string]interface{}{
		"top_line":   s.topLine,
		"bot_line":   s.botLine,
		"cur_line":   s.curLine,
		"line_count": s.lineCount,
	})
}

func (s *Screen) handleWinViewportMargins(args []interface{}) {
	if len(args) < 1 {
		return
	}

	marginArgs, ok := args[0].([]interface{})
	if !ok || len(marginArgs) < 7 {
		return
	}

	grid := utils.ReflectToInt(marginArgs[0])
	if grid != 1 {
		return // Only care about main grid
	}

	// win := utils.ReflectToInt(marginArgs[1]) // Window ID
	s.margins[0] = utils.ReflectToInt(marginArgs[2]) // top
	s.margins[1] = utils.ReflectToInt(marginArgs[3]) // bottom
	s.margins[2] = utils.ReflectToInt(marginArgs[4]) // left
	s.margins[3] = utils.ReflectToInt(marginArgs[5]) // right

}

func (s *Screen) gridResize(gridId int, width int, height int) {
	if gridId > 2 {
		utils.Log(fmt.Sprintf("gridResize gridId: %d, width: %d, height: %d", gridId, width, height))
	}
	grid, exists := s.Grids[gridId]
	if !exists {
		grid = &Grid{
			ID:     gridId,
			Width:  0,
			Height: 0,
			Cells:  nil,
		}
		s.Grids[gridId] = grid
	}

	grid.Width = width
	grid.Height = height

	// Resize the grid cells
	newCells := make([][]*Cell, height)
	for i := range newCells {
		newCells[i] = make([]*Cell, width)
		for j := range newCells[i] {
			// Copy existing cell if available
			if i < len(grid.Cells) && j < len(grid.Cells[i]) && grid.Cells[i][j] != nil {
				newCells[i][j] = &Cell{
					Char:      grid.Cells[i][j].Char,
					Highlight: grid.Cells[i][j].Highlight,
				}
			} else {
				newCells[i][j] = &Cell{
					Char:      " ",
					Highlight: 0,
				}
			}
		}
	}

	grid.Cells = newCells
	// Update window dimensions if this grid is associated with a window
	s.windowsMu.RLock()
	if winId, exists := s.GridToWindow[gridId]; exists {
		if window, exists := s.Windows[winId]; exists {
			window.Width = width
			window.Height = height

			// If this is a small 1x1 window, it's probably not a completion window
			if width == 1 && height == 1 {
				window.IsPopupmenu = false
			}

			utils.Log(fmt.Sprintf("Updated window %d dimensions to %dx%d", winId, width, height))
		}
	}
	s.windowsMu.RUnlock()
	// s.scheduleRender()
	//TODO: probably dont wanna recalculate everything since only the fractions will change
}

func (s *Screen) gridCursorGoto(gridId int, row int, col int) {
	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}

	grid.Cursor.Row = row
	grid.Cursor.Col = col

	s.ActiveGrid = gridId
	s.ActiveWindow = s.GridToWindow[gridId]

	utils.Log("gridCursorGoto updated active Window to ", s.ActiveWindow)

	UpdateCursor(s.ctx, CursorMoveEvent{
		Row:            uint64(row),
		Col:            uint64(col),
		TopLine:        uint64(s.botLine),
		BottomLine:     uint64(s.topLine),
		ActiveWindowId: s.ActiveWindow,
	})
}

func (s *Screen) defaultColorsSet(fg int, bg int, sp int) {
	if fg >= 0 {
		s.DefaultFg = fg
	} else {
		s.DefaultFg = 0xffffff
	}

	if bg >= 0 {
		s.DefaultBg = bg
	} else {
		s.DefaultBg = 0x000000
	}

	if sp >= 0 {
		s.DefaultSp = sp
	} else {
		s.DefaultSp = s.DefaultFg
	}

	// Update default highlight
	s.Highlights[0] = &Highlight{
		Foreground: s.DefaultFg,
		Background: s.DefaultBg,
		Special:    s.DefaultSp,
	}

	s.scheduleRender()
}

func (s *Screen) sendInitialHighlights(optionalData ...interface{}) {
	highlightUpdates := make([]map[string]interface{}, 0)

	s.highlightsMu.RLock()
	for id, highlight := range s.Highlights {
		highlightDef := map[string]interface{}{
			"id":            id,
			"fg":            highlight.fgHex(),
			"bg":            highlight.bgHex(),
			"bold":          highlight.Bold,
			"italic":        highlight.Italic,
			"underline":     highlight.Underline,
			"undercurl":     highlight.Undercurl,
			"strikethrough": highlight.Strikethrough,
			"reverse":       highlight.Reverse,
		}

		highlightUpdates = append(highlightUpdates, highlightDef)
	}
	s.highlightsMu.RUnlock()

	if len(highlightUpdates) > 0 {
		Runtime.EventsEmit(s.ctx, "highlight_defined", highlightUpdates)
	}
}

func (s *Screen) hlAttrDefine(args []interface{}) {
	highlightUpdates := make([]map[string]interface{}, 0)

	for _, attr := range args {
		attrData := attr.([]interface{})
		id := utils.ReflectToInt(attrData[0])
		rgbAttrs := attrData[1].(map[string]interface{})

		highlight := &Highlight{}

		if fg, ok := rgbAttrs["foreground"]; ok {
			highlight.Foreground = utils.ReflectToInt(fg)
		}

		if bg, ok := rgbAttrs["background"]; ok {
			highlight.Background = utils.ReflectToInt(bg)
		}

		if sp, ok := rgbAttrs["special"]; ok {
			highlight.Special = utils.ReflectToInt(sp)
		}

		if reverse, ok := rgbAttrs["reverse"]; ok {
			highlight.Reverse = reverse.(bool)
		}

		if italic, ok := rgbAttrs["italic"]; ok {
			highlight.Italic = italic.(bool)
		}

		if bold, ok := rgbAttrs["bold"]; ok {
			highlight.Bold = bold.(bool)
		}

		if underline, ok := rgbAttrs["underline"]; ok {
			highlight.Underline = underline.(bool)
		}

		if undercurl, ok := rgbAttrs["undercurl"]; ok {
			highlight.Undercurl = undercurl.(bool)
		}

		if strikethrough, ok := rgbAttrs["strikethrough"]; ok {
			highlight.Strikethrough = strikethrough.(bool)
		}

		s.highlightsMu.Lock()
		s.Highlights[id] = highlight
		s.highlightsMu.Unlock()

		// Create a map with highlight properties to send to frontend
		highlightDef := map[string]interface{}{
			"id":            id,
			"fg":            highlight.fgHex(),
			"bg":            highlight.bgHex(),
			"bold":          highlight.Bold,
			"italic":        highlight.Italic,
			"underline":     highlight.Underline,
			"undercurl":     highlight.Undercurl,
			"strikethrough": highlight.Strikethrough,
			"reverse":       highlight.Reverse,
		}

		highlightUpdates = append(highlightUpdates, highlightDef)
	}

	// Emit highlight definitions to frontend
	if len(highlightUpdates) > 0 {
		Runtime.EventsEmit(s.ctx, "highlight_defined", highlightUpdates)
	}

	s.scheduleRender()
}

func (s *Screen) modeChange(mode string) {
	s.Mode = mode
	Runtime.EventsEmit(s.ctx, "mode-changed", mode)
}

func (s *Screen) winClose(args []interface{}) {
	utils.Log("winClose", args)
	if len(args) < 1 {
		return
	}

	gridId := utils.ReflectToInt(args[0])
	utils.Log("winClose", args)

	// Find the grid associated with this window
	s.windowsMu.RLock()
	for grid, winId := range s.GridToWindow {
		if grid == gridId {
			if s.Windows[winId].IsFloating() {
				Runtime.EventsEmit(s.ctx, "floating_window_closed", winId)
			}
			delete(s.GridToWindow, grid)
			delete(s.Windows, winId)
			utils.Log(fmt.Sprintf("Window %d closed (grid %d)", winId, gridId))
			utils.Log(fmt.Sprintf("winPos got %d windows now", len(s.Windows)))
			break
		}
	}
	s.windowsMu.RUnlock()
}

func (s *Screen) winPos(args []interface{}) {
	if len(args) < 6 {
		return
	}

	gridId := utils.ReflectToInt(args[0])
	nwindow := args[1].(nvim.Window)
	winId, _ := strconv.Atoi(strings.Split(nwindow.String(), ":")[1])
	row := utils.ReflectToInt(args[2])
	col := utils.ReflectToInt(args[3])
	width := utils.ReflectToInt(args[4])
	height := utils.ReflectToInt(args[5])

	s.windowsMu.RLock()
	window, exists := s.Windows[winId]
	if !exists {
		window = NewWindow(winId, s.Grids[gridId])

		s.Windows[winId] = window
		s.GridToWindow[gridId] = winId
		Runtime.EventsEmit(s.ctx, "window_opened")
		utils.Log(fmt.Sprintf("winPos spawned with id=%d got %d windows now", winId, len(s.Windows)))
	}
	window.Width = width
	window.Height = height
	window.StartRow = row
	window.StartCol = col
	utils.Log(fmt.Sprintf("winPos id=%d StartRow=%d StartCol=%d Width=%d Height=%d", winId, row, col, width, height))
	s.windowsMu.RUnlock()

	// Handle window positioning
	if _, exists := s.Grids[gridId]; !exists {
		// Create a new grid for this window
		s.gridResize(gridId, width, height)
	}
}

func (s *Screen) updateLayout() {
	layout := s.CalculateGridLayout()
	layout.ActiveWindowId = s.GridToWindow[s.ActiveGrid]
	Runtime.EventsEmit(s.ctx, "layout-updated", layout)
}

func (s *Screen) winFloatPos(args []interface{}) {
	utils.Log("winFloatPos: ", args)
	if len(args) < 8 {
		utils.Log(fmt.Sprintf("winFloatPos not enough args %d", len(args)))
		return
	}

	gridId := utils.ReflectToInt(args[0])
	utils.Log(fmt.Sprintf("winFloatPos computing args for gridId %d", gridId), args)
	nwindow := args[1].(nvim.Window)
	winId, _ := strconv.Atoi(strings.Split(nwindow.String(), ":")[1])
	anchor := args[2].(string)
	anchorGrid := utils.ReflectToInt(args[3])
	anchorRow := utils.ReflectToFloat(args[4])
	anchorCol := utils.ReflectToFloat(args[5])
	focusable := args[6].(bool)
	zIndex := utils.ReflectToInt(args[7])
	utils.Log(fmt.Sprintf("winFloatPos winId: %d row: %f col %f", winId, anchorRow, anchorCol), args)

	// Update window tracking
	s.windowsMu.Lock()
	defer s.windowsMu.Unlock()

	window, exists := s.Windows[winId]
	if !exists {
		utils.Log(fmt.Sprintf("winFloatPos adding winId %d", winId))
		window = NewWindow(winId, s.Grids[gridId])
		s.Windows[winId] = window
	} else {
		utils.Log(fmt.Sprintf("winFloatPos already had winId %d", winId))
	}

	existingGrid := s.Grids[gridId]
	if existingGrid != nil {
		window.Width = existingGrid.Width
		window.Height = existingGrid.Height
	}

	// Update window properties
	window.Grid.ID = gridId
	window.Type = "floating"
	window.Anchor = anchor
	window.AnchorGrid = anchorGrid
	window.StartRow = int(anchorRow)
	window.StartCol = int(anchorCol)
	window.Focusable = focusable
	window.ZIndex = zIndex

	// Check if this might be a completion window
	// Completion windows are typically floating windows with specific characteristics
	if window.Width > 1 && window.Height > 1 {
		// This is a heuristic - we might need to refine this
		window.IsPopupmenu = true
	}

	// Update grid to window mapping
	s.GridToWindow[gridId] = winId

	utils.Log(fmt.Sprintf("Floating window %d anchored at grid %d (%f,%f) with z-index %d",
		winId, anchorGrid, anchorRow, anchorCol, zIndex))
}

func (s *Screen) scheduleRender() {
	if !s.PendingRender {
		s.PendingRender = true
		// Use goroutine to simulate setImmediate behavior
		go func() {
			// Small delay to batch updates
			time.Sleep(time.Millisecond * 5)
			s.render()
			s.PendingRender = false
		}()
	}
}

func (s *Screen) render() {
	s.EmitFloatingWindows()
	s.updateLayout()
	for winId, window := range s.Windows {
		if !window.Dirty {
			continue
		}
		window.Dirty = false
		Runtime.EventsEmit(s.ctx, "content-updated", winId, s.optimizeGrid(window.Grid))
	}
}

func (s *Screen) optimizeGrid(grid *Grid) [][]*Cell {
	cursor := s.Grids[grid.ID].Cursor
	optimizedGrid := make([][]*Cell, len(grid.Cells))

	for i, row := range grid.Cells {
		optimizedGrid[i] = make([]*Cell, 0)
		if len(row) == 0 {
			continue
		}

		if len(row) > 0 {
			firstCell := row[0]
			lastHl := firstCell.Highlight
			currentToken := Cell{
				Char:      "",
				Highlight: lastHl,
				Dirty:     firstCell.Dirty,
			}

			for j, cell := range row {
				isCursor := cursor.Row == i && cursor.Col == j

				// If we've reached the cursor, save current token and start a cursor token
				if isCursor {
					// If we have accumulated characters, store the current token first
					if len(currentToken.Char) > 0 {
						tokenCopy := currentToken
						optimizedGrid[i] = append(optimizedGrid[i], &tokenCopy)
					}

					// Create the cursor token
					cursorToken := Cell{
						Char:      cell.Char,
						Highlight: cell.Highlight,
						Dirty:     true,
						Classes:   "cursor",
					}

					optimizedGrid[i] = append(optimizedGrid[i], &cursorToken)

					// Start a new token for characters after cursor
					currentToken = Cell{
						Char:      "",
						Highlight: cell.Highlight,
						Dirty:     cell.Dirty,
					}
					lastHl = cell.Highlight
				} else if lastHl == cell.Highlight || cell.Char == " " {
					// Same highlight or space, append to current token
					currentToken.Char += cell.Char
					currentToken.Dirty = currentToken.Dirty || cell.Dirty
				} else {
					// Different highlight, store current token and start a new one
					if len(currentToken.Char) > 0 {
						tokenCopy := currentToken
						optimizedGrid[i] = append(optimizedGrid[i], &tokenCopy)
					}

					// Start new token
					lastHl = cell.Highlight
					currentToken = Cell{
						Char:      cell.Char,
						Highlight: cell.Highlight,
						Dirty:     cell.Dirty,
					}
				}
			}

			// Don't forget to add the last token from the row
			if len(currentToken.Char) > 0 {
				tokenCopy := currentToken
				optimizedGrid[i] = append(optimizedGrid[i], &tokenCopy)
			}
		}
	}
	for _, optimizedRow := range optimizedGrid {
		optimizedRow[len(optimizedRow)-1].Char = strings.TrimRight(optimizedRow[len(optimizedRow)-1].Char, " ")
	}

	return optimizedGrid
}

func sanitize(s string) string {
	s = strings.Replace(s, " ", `&nbsp;`, -1)
	s = strings.Replace(s, "\t", `&nbsp;`, -1)
	s = strings.Replace(s, "<", `&lt;`, -1)
	s = strings.Replace(s, ">", `&gt;`, -1)

	return s
}

func (s *Screen) handleCmdlineShow(args []interface{}) {
	arg := args[0].([]interface{})

	content := ""
	contentChunks := arg[0].([]interface{})
	for _, e := range contentChunks {
		a := e.([]interface{})

		if len(a) < 2 {
			// content += a[0].(string)
			content += strings.Replace(a[0].(string), "\t", " ", -1)
		} else {
			if len(contentChunks) == 1 {
				// content += a[1].(string)
				content += strings.Replace(a[1].(string), "\t", " ", -1)
			} else {
				content +=
					sanitize(a[1].(string))
			}
		}
	}
	// content := arg[0].([]interface{})[0].([]interface{})[1].(string)

	pos := util.ReflectToInt(arg[1])
	firstc := arg[2].(string)
	prompt := arg[3].(string)
	indent := util.ReflectToInt(arg[4])
	// level := util.ReflectToInt(arg[5])
	// fmt.Println("cmdline show", content, pos, firstc, prompt, indent, level)

	Runtime.EventsEmit(s.ctx, "cmdline_show", map[string]interface{}{
		"content": content,
		"pos":     pos,
		"firstc":  firstc,
		"prompt":  prompt,
		"indent":  indent,
	})
}

// handleCmdlinePos processes the cmdline_pos event
func (s *Screen) handleCmdlinePos(args []interface{}) {
	cmdlineArgs := args[0].([]interface{})
	if len(cmdlineArgs) < 2 {
		return
	}

	pos := utils.ReflectToInt(cmdlineArgs[0])
	level := utils.ReflectToInt(cmdlineArgs[1])

	Runtime.EventsEmit(s.ctx, "cmdline_pos", map[string]interface{}{
		"pos":   pos,
		"level": level,
	})
}

func (s *Screen) EmitFloatingWindows() {
	s.windowsMu.RLock()
	defer s.windowsMu.RUnlock()

	var floatingWindows []map[string]interface{}

	for winId, window := range s.Windows {
		if !window.IsFloating() {
			continue
		}
		// Get the associated grid
		grid, exists := s.Grids[window.Grid.ID]
		if !exists {
			utils.Log(fmt.Sprintf("emitfloat got no grid for %d", window.Grid.ID))
			continue
		} else {
			utils.Log(fmt.Sprintf("emitfloat found grid for %d", window.Grid.ID))
		}

		str := strings.TrimSpace(grid.toString())
		// seems arbitrary, but theres some weird floating windows with bs content
		// idk wtf they are and i dont care. i want them gone. cant imagine they could be important with 3 chars
		if len(str) < 4 {
			continue
		} else {
			utils.Log(fmt.Sprintf("EmitFloatingWindows %d longer than 3 chars:%s", len(str), str))
		}
		anchorWindow, _ := s.GridToWindow[window.AnchorGrid]

		// Create a window info object
		windowInfo := map[string]interface{}{
			"id":           winId,
			"grid_id":      window.Grid.ID,
			"anchorWindow": anchorWindow,
			"anchor":       window.Anchor,
			"row":          window.StartRow,
			"col":          window.StartCol,
			"width":        window.Width,
			"height":       window.Height,
			"z_index":      window.ZIndex,
			"focusable":    window.Focusable,
			"is_popup":     window.IsPopupmenu,
			"grid":         grid.toHexGrid().Cells,
		}

		floatingWindows = append(floatingWindows, windowInfo)
	}

	if len(floatingWindows) > 0 {
		Runtime.EventsEmit(s.ctx, "floating_windows", floatingWindows)
	}
}
