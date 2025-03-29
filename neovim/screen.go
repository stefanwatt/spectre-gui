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
	Char      string `json:"char"`
	Highlight int    `json:"highlight"`
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
			utils.Log("event not ok")
			continue
		}
		utils.Log(fmt.Sprintf("event: %s", event))
		args := update[1:]

		switch event {
		case "grid_clear":
			s.gridClear(args)
		case "grid_line":
			utils.Log("grid_line")
			s.gridLine(args)
		case "grid_scroll":
			s.gridScroll(args)
		case "grid_resize":
			s.gridResize(args)
		case "grid_cursor_goto":
			s.gridCursorGoto(args)
		case "flush":
			s.flush()
		case "win_viewport":
			s.handleWinViewport(args)
		case "win_viewport_margins":
			s.handleWinViewportMargins(args)
		case "default_colors_set":
			s.defaultColorsSet(args)
		case "hl_attr_define":
			s.hlAttrDefine(args)
		case "mode_change":
			s.modeChange(args)
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

func (s *Screen) gridScroll(args []interface{}) {
	if len(args) < 1 {
		return
	}
	scrollArgs, ok := args[0].([]interface{})
	if !ok || len(scrollArgs) < 7 {
		return
	}

	gridId := utils.ReflectToInt(scrollArgs[0])
	top := utils.ReflectToInt(scrollArgs[1])
	bot := utils.ReflectToInt(scrollArgs[2])
	left := utils.ReflectToInt(scrollArgs[3])
	right := utils.ReflectToInt(scrollArgs[4])
	rows := utils.ReflectToInt(scrollArgs[5])
	cols := utils.ReflectToInt(scrollArgs[6])

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

		if direction > 0 {
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

	s.scheduleRender()
}

func (s *Screen) gridLine(args []interface{}) {
	if len(args) < 1 {
		return
	}
	gridArgs, ok := args[0].([]interface{})
	if !ok || len(gridArgs) < 4 {
		return
	}

	gridId := utils.ReflectToInt(gridArgs[0])
	utils.Log(fmt.Sprintf("grid_line event for grid#%d", gridId))
	row := utils.ReflectToInt(gridArgs[1])
	col := utils.ReflectToInt(gridArgs[2])

	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}

	// Debug output for rows we're interested in
	if row >= 7 && row <= 10 {
		utils.Log(fmt.Sprintf("grid_line: grid=%d, row=%d, col=%d", gridId, row, col))
	}

	// Check if this row is within the grid
	if row >= grid.Height || col >= grid.Width {
		utils.Log(fmt.Sprintf("Row %d or col %d out of bounds for grid %d (max: %d,%d)",
			row, col, gridId, grid.Height-1, grid.Width-1))
		return
	}

	cells := gridArgs[3].([]interface{})

	currentCol := col
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
		}

		for i := 0; i < repeat && currentCol < grid.Width; i++ {
			if row < len(grid.Cells) && currentCol < len(grid.Cells[row]) {
				grid.Cells[row][currentCol] = &Cell{
					Char:      char,
					Highlight: hl,
				}
				currentCol++
			}
		}
	}
	s.scheduleRender()
}

func (s *Screen) flush() {
	s.render()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Screen) gridClear(args []interface{}) {
	for _, arg := range args {
		gridId := utils.ReflectToInt(arg.([]interface{})[0])
		grid, exists := s.Grids[gridId]
		if !exists {
			continue
		}

		// Clear all content
		for i := range grid.Cells {
			for j := range grid.Cells[i] {
				grid.Cells[i][j] = &Cell{
					Char:      " ",
					Highlight: 0,
				}
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

func (s *Screen) gridResize(args []interface{}) {
	if len(args) < 1 {
		return
	}
	resizeArgs, ok := args[0].([]interface{})
	if !ok || len(resizeArgs) < 3 {
		utils.Log("Invalid grid_resize args format")
		return
	}

	gridId := utils.ReflectToInt(resizeArgs[0])
	width := utils.ReflectToInt(resizeArgs[1])
	height := utils.ReflectToInt(resizeArgs[2])

	utils.Log(fmt.Sprintf("grid_resize: grid=%d, width=%d, height=%d", gridId, width, height))

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

func (s *Screen) gridCursorGoto(args []interface{}) {
	if len(args) < 1 {
		return
	}

	gotoArgs := args[0].([]interface{})
	if len(gotoArgs) < 3 {
		return
	}

	gridId := utils.ReflectToInt(gotoArgs[0])
	row := utils.ReflectToInt(gotoArgs[1])
	col := utils.ReflectToInt(gotoArgs[2])

	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}

	grid.Cursor.Row = row
	grid.Cursor.Col = col
	s.ActiveGrid = gridId

	s.scheduleRender()
}

func (s *Screen) defaultColorsSet(args []interface{}) {
	if len(args) < 3 {
		return
	}

	fg := utils.ReflectToInt(args[0])
	bg := utils.ReflectToInt(args[1])
	sp := utils.ReflectToInt(args[2])

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

func (s *Screen) modeChange(args []interface{}) {
	if len(args) < 2 {
		return
	}

	mode := args[0].(string)
	// modeIdx := utils.ReflectToInt(args[1])

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
		s.gridResize([]interface{}{
			[]interface{}{gridId, width, height},
		})
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
		s.gridResize([]interface{}{
			[]interface{}{gridId, 10, 5}, // Default width=10, height=5
		})
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
	// Render all visible grids into the buffer
	utils.Log(fmt.Sprintf("rendering %d grids", len(s.Grids)))
	if len(s.Grids) == 1 {
		utils.Log("grid 1 ", s.Grids[0])
	}
	for gridId, grid := range s.Grids {
		utils.Log(fmt.Sprintf("grid#%d", gridId), grid.toString())
		if gridId != 2 {
			continue
		}

		for row := 0; row < grid.Height && row < len(s.Content); row++ {
			for col := 0; col < grid.Width && col < len(s.Content[row]); col++ {
				if row < len(grid.Cells) && col < len(grid.Cells[row]) {
					if grid.Cells[row][col] != nil {
						s.Content[row][col] = grid.Cells[row][col]
					}
				}
			}
		}
	}
	Runtime.EventsEmit(s.ctx, "flush", s.Content)
}
