package neovim

import (
	"context"
	"fmt"
	"math"
	"nvim-gui/utils"
	"time"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Grid struct {
	ID     int
	Width  int
	Height int
	Cells  [][]*Cell
	Cursor struct {
		Row int
		Col int
	}
}

func (g *Grid) toString() string {
	result := ""
	for _, row := range g.Cells {
		for _, cell := range row {
			result += cell.Char
		}
		result += "\n"
	}
	return result
}

type Cell struct {
	Char       string `json:"char"`
	Highlight  int
	Foreground string `json:"fg"`
	Background string `json:"bg"`
	Dirty      bool
	Classes    string `json:"classes"`
}

func (c *Cell) Equals(other *Cell) bool {
	if c == nil || other == nil {
		return c == other
	}
	return c.Char == other.Char && c.Foreground == other.Foreground && c.Background == other.Background
}

type Screen struct {
	margins       []int
	rows          int
	cols          int
	topLine       int
	botLine       int
	curLine       int
	lineCount     int
	scrollDelta   float64
	ctx           context.Context
	Content       [][]*Cell
	Grids         map[int]*Grid
	Highlights    map[int]*Highlight
	DefaultFg     int
	DefaultBg     int
	DefaultSp     int
	ActiveGrid    int
	Mode          string
	PendingRender bool
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
		Grids:         grids,
		Highlights:    highlights,
		DefaultFg:     0xffffff,
		DefaultBg:     0x000000,
		DefaultSp:     0xffffff,
		ActiveGrid:    2,
		Mode:          "normal",
		PendingRender: false,
		Content:       content,
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
			s.winPos(args)
		case "win_float_pos":
			s.winFloatPos(args)
		case "cmdline_show":
			Runtime.EventsEmit(s.ctx, "cmdline_show")
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

	utils.Log(fmt.Sprintf("gridScroll id:%d, top:%d, bot:%d, left:%d, right:%d, rows:%d, cols:%d",
		gridId, top, bot, left, right, rows, cols))

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
		s.scheduleRender()
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
				highlight := s.Highlights[hl]
				grid.Cells[row][currentCol] = &Cell{
					Char:       char,
					Highlight:  hl,
					Foreground: highlight.fgHex(),
					Background: highlight.bgHex(),
				}
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
	s.scheduleRender()
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

	utils.Log(fmt.Sprintf("Viewport: topLine=%d, botLine=%d, curLine=%d, lineCount=%d, scrollDelta=%f",
		s.topLine, s.botLine, s.curLine, s.lineCount, s.scrollDelta))

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

	utils.Log(fmt.Sprintf("Viewport margins: top=%d, bottom=%d, left=%d, right=%d",
		s.margins[0], s.margins[1], s.margins[2], s.margins[3]))
}

func (s *Screen) gridResize(gridId int, width int, height int) {
	// Get or create the grid
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
	s.scheduleRender()
}

func (s *Screen) gridCursorGoto(gridId int, row int, col int) {
	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}

	grid.Cursor.Row = row
	grid.Cursor.Col = col

	s.ActiveGrid = gridId

	UpdateCursor(s.ctx, CursorMoveEvent{
		Row:        uint64(row),
		Col:        uint64(col),
		TopLine:    uint64(s.botLine),
		BottomLine: uint64(s.topLine),
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

func (s *Screen) hlAttrDefine(args []interface{}) {
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

		s.Highlights[id] = highlight
	}

	s.scheduleRender()
}

func (s *Screen) modeChange(mode string) {
	s.Mode = mode
	s.scheduleRender()
}

func (s *Screen) winPos(args []interface{}) {
	if len(args) < 6 {
		return
	}

	gridId := utils.ReflectToInt(args[0])
	// win := args[1]
	// row := utils.ReflectToInt(args[2])
	// col := utils.ReflectToInt(args[3])
	width := utils.ReflectToInt(args[4])
	height := utils.ReflectToInt(args[5])

	// Handle window positioning
	if _, exists := s.Grids[gridId]; !exists {
		// Create a new grid for this window
		s.gridResize(gridId, width, height)
	}
}

func (s *Screen) winFloatPos(args []interface{}) {
	if len(args) < 8 {
		return
	}

	gridId := utils.ReflectToInt(args[0])
	// win := args[1]
	// anchor := args[2].(string)
	// anchorGrid := utils.ReflectToInt(args[3])
	// anchorRow := utils.ReflectToInt(args[4])
	// anchorCol := utils.ReflectToInt(args[5])
	// focusable := args[6].(bool)
	// zIndex := utils.ReflectToInt(args[7])

	// Handle floating window positioning
	if _, exists := s.Grids[gridId]; !exists {
		// Create a new grid for this floating window with default size
		s.gridResize(gridId, 10, 5)
	}
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
	var grid *Grid
	for gridId, g := range s.Grids {
		if gridId != 2 {
			continue
		}
		grid = g
	}
	if grid == nil {
		return
	}

	for row := 0; row < grid.Height && row < len(s.Content); row++ {
		for col := 0; col < grid.Width && col < len(s.Content[row]); col++ {
			if row < len(grid.Cells) && col < len(grid.Cells[row]) {
				if grid.Cells[row][col] != nil {
					updatedCell := grid.Cells[row][col]
					if !updatedCell.Equals(s.Content[row][col]) {
						updatedCell.Dirty = true
					}
					s.Content[row][col] = updatedCell
				}
			}
		}
	}
	// updates := s.createSparseUpdates()
	// Runtime.EventsEmit(s.ctx, "flush", updates)
	Runtime.EventsEmit(s.ctx, "flush", s.optimizeGrid())
}

func (s *Screen) optimizeGrid() [][]*Cell {

	optimizedGrid := make([][]*Cell, len(s.Content))
	for i, row := range s.Content {
		optimizedGrid[i] = make([]*Cell, 0) // Initialize with empty slice, we'll append
		if len(row) == 0 {
			continue // Skip empty rows
		}

		firstCell := row[0]
		lastHl := firstCell.Highlight
		currentToken := Cell{
			Char:       "",
			Highlight:  lastHl,
			Background: firstCell.Background,
			Foreground: firstCell.Foreground,
			Dirty:      firstCell.Dirty,
		}

		for _, cell := range row {
			if lastHl == cell.Highlight || cell.Char == " " {
				currentToken.Char += cell.Char
				currentToken.Dirty = currentToken.Dirty || cell.Dirty
			} else {
				// Different highlight, store current token and start a new one
				tokenCopy := currentToken // Copy to avoid reference issues
				optimizedGrid[i] = append(optimizedGrid[i], &tokenCopy)

				// Start new token
				lastHl = cell.Highlight
				currentToken = Cell{
					Char:       cell.Char,
					Highlight:  cell.Highlight,
					Background: cell.Background,
					Foreground: cell.Foreground,
					Dirty:      cell.Dirty,
				}
			}
		}

		// Don't forget to add the last token from the row
		tokenCopy := currentToken
		optimizedGrid[i] = append(optimizedGrid[i], &tokenCopy)
	}
	return optimizedGrid
}

func (s *Screen) createSparseUpdates() [][]*Cell {
	updates := make([][]*Cell, len(s.Content))
	hasChanges := false

	// Compare each cell and only include changed ones
	for i, row := range s.Content {
		rowHasChanges := false
		updates[i] = make([]*Cell, len(row))

		for j, cell := range row {
			if cell.Dirty {
				updates[i][j] = cell
				cell.Dirty = false
				rowHasChanges = true
				hasChanges = true
			} else {
				updates[i][j] = nil // Unchanged cell
			}
		}

		if !rowHasChanges {
			updates[i] = nil // Entire row unchanged
		}
	}

	// Only return updates if there are changes
	if hasChanges {
		return updates
	}
	return nil
}
