package neovim

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"nvim-gui/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akiyosi/goneovim/util"
	"github.com/neovim/go-client/nvim"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Screen struct {
	margins              []int
	Height               int //in number of cells
	Width                int //in number of cells
	topLine              int
	botLine              int
	curLine              int
	lineCount            int
	scrollDelta          float64
	ctx                  context.Context
	app                  *application.App // Reference to Wails app for event emission
	Grids                map[int]*Grid
	Highlights           map[int]*Highlight
	Windows              map[int]*Window // Map of window IDs to Window objects
	GridToWindow         map[int]int     // Map of grid IDs to window IDs
	DefaultFg            int
	DefaultBg            int
	DefaultSp            int
	ActiveGrid           int
	ActiveWindow         int
	Mode                 string
	PendingRender        bool
	highlightsMu         sync.RWMutex // Mutex for Highlights map
	windowsMu            sync.RWMutex // Mutex for Windows map
	tableMetadata        map[int][]TableMeta // bufNr -> []TableMeta
	tableMetadataMu      sync.RWMutex
	imageMetadata        map[int][]ImageMeta // bufNr -> []ImageMeta
	imageMetadataMu      sync.RWMutex
	headingMetadata      map[int]map[int]int // bufNr -> line -> heading level
	headingMetadataMu    sync.RWMutex
	taskMetadata         map[int]map[int]bool // bufNr -> line -> checked
	taskMetadataMu       sync.RWMutex
	codeBlockMetadata    map[int][]CodeBlockMeta // bufNr -> []CodeBlockMeta
	codeBlockMetadataMu  sync.RWMutex
	inlineCodeMetadata   map[int]map[int][]ColRange // bufNr -> line -> []ColRange
	inlineCodeMetadataMu sync.RWMutex
	ColorColumns         []int  // columns where colorcolumn should render (e.g. [80])
	ColorColumnColor     string // hex color e.g. "#2a2a3a"
	CursorLineEnabled    bool
	CursorLineColor      string // hex color e.g. "#2a2a3a"
}

// emitEvent safely emits an event, checking if app is initialized
func (s *Screen) emitEvent(eventName string, data interface{}) {
	if s.app == nil {
		utils.Log(fmt.Sprintf("Warning: Cannot emit event %s - app not initialized", eventName))
		return
	}
	s.app.Event.Emit(eventName, data)
}

func NewScreen(ctx context.Context, cols int, rows int, app *application.App) *Screen {

	// Create grids map and add default grid
	grids := make(map[int]*Grid)
	grids[1] = NewGrid(rows, cols)

	// Create default highlight
	highlights := make(map[int]*Highlight)
	highlights[0] = &Highlight{
		Foreground: 0xffffff,
		Background: 0x000000,
		Special:    0xffffff,
	}

	return &Screen{
		ctx:                ctx,
		app:                app,
		Width:              cols,
		Height:             rows,
		Grids:              grids,
		Highlights:         highlights,
		Windows:            make(map[int]*Window),
		GridToWindow:       make(map[int]int),
		DefaultFg:          0xffffff,
		DefaultBg:          0x000000,
		DefaultSp:          0xffffff,
		ActiveGrid:         2,
		Mode:               "normal",
		PendingRender:      false,
		margins:            make([]int, 4), // Initialize margins slice
		tableMetadata:      make(map[int][]TableMeta),
		imageMetadata:      make(map[int][]ImageMeta),
		headingMetadata:    make(map[int]map[int]int),
		taskMetadata:       make(map[int]map[int]bool),
		codeBlockMetadata:  make(map[int][]CodeBlockMeta),
		inlineCodeMetadata: make(map[int]map[int][]ColRange),
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
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				s.winPos(gridArgs)
			}
		case "win_float_pos":
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				s.winFloatPos(gridArgs)
			}
		case "win_external_pos":
			for _, arg := range args {
				extArgs := arg.([]interface{})
				s.winExternalPos(extArgs)
			}
		case "win_close":
			utils.Log("redraw win_close args:", args)
			utils.Log("redraw win_close s.Grids", s.Grids)
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				s.winClose(gridArgs)
			}

		case "win_hide":
			utils.Log("redraw win_hide args:", args)
			utils.Log("redraw win_hide s.Grids", s.Grids)
			for _, arg := range args {
				gridArgs := arg.([]interface{})
				s.winHide(gridArgs)
			}
		case "cmdline_show":
			s.handleCmdlineShow(args)
		case "cmdline_pos":
			s.handleCmdlinePos(args)
		case "cmdline_hide":
			s.emitEvent("cmdline_hide", struct{}{})
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
	startRow := top
	endRow := bot
	if rows < 0 {
		startRow = top + (-rows)
	} else {
		endRow = bot - rows
	}
	for r := startRow; r < endRow; r++ {
		if r >= 0 && r < len(grid.DirtyRows) {
			grid.DirtyRows[r] = true
		}
	}
	// Mark the window dirty so render() emits content-updated
	if winId, exists := s.GridToWindow[gridId]; exists {
		s.windowsMu.RLock()
		if window, exists := s.Windows[winId]; exists {
			window.Dirty = true
		}
		s.windowsMu.RUnlock()
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
		window, exists := s.Windows[winId]
		if exists {
			window.Dirty = true
		}
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

				// Ensure hl is valid
				s.highlightsMu.RLock()
				if _, exists := s.Highlights[hl]; !exists {
					hl = 0 // Default to 0 if the highlight ID doesn't exist
				}
				highlight := s.Highlights[hl]
				s.highlightsMu.RUnlock()
				hlStr := mapClassesString(highlight.getClasses())
				effectiveHlIdsMu.Lock()
				effectiveHlId, existsHlId := effectiveHlIds[hlStr]
				effectiveHlIdsMu.Unlock()
				if !existsHlId {
					classesMap := make(map[string]bool)
					for _, class := range highlight.getClasses() {
						classesMap[class] = true
					}
					newCell := Cell{
						Char:      char,
						Highlight: hl,
						Classes:   classesMap,
					}
					grid.Cells[row][currentCol] = &newCell
					currentCol++
					continue // Skip the rest of the loop for this cell
				}

				// Retrieve pre-calculated classes using the effective ID
				classes, exists := idClasses[effectiveHlId]
				if !exists {
					// Log error and use default classes
					utils.Log(fmt.Sprintf("Error: Classes not found for effectiveHlId=%d (original hl=%d). Using default.", effectiveHlId, hl))
					classes = idClasses[0] // Use default classes
				}

				classesMap := make(map[string]bool)
				for _, class := range classes {
					if class == "fg-4" || class == "fg-15" {
						utils.Log(fmt.Sprintf("gridLine found weird turquoise fg color char=%s highlight.fgHex=%s", char, highlight.fgHex()))
					}
					classesMap[class] = true
				}
				newCell := Cell{
					Char:      char,
					Highlight: hl,
					Classes:   classesMap,
				}
				grid.Cells[row][currentCol] = &newCell
				currentCol++
			}
		}
	}

	grid.DirtyRows[row] = true
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
	for _, arg := range args {
		viewportArgs, ok := arg.([]interface{})
		if !ok || len(viewportArgs) < 8 {
			continue
		}

		gridId := utils.ReflectToInt(viewportArgs[0])
		topLine := utils.ReflectToInt(viewportArgs[2])
		botLine := utils.ReflectToInt(viewportArgs[3])
		curLine := utils.ReflectToInt(viewportArgs[4])
		lineCount := utils.ReflectToInt(viewportArgs[6])

		// Store topline on the grid for line number calculation
		if grid, exists := s.Grids[gridId]; exists {
			if grid.TopLine != topLine {
				grid.TopLine = topLine
				// TopLine change means line numbers shift — mark window dirty
				if winId, exists := s.GridToWindow[gridId]; exists {
					s.windowsMu.RLock()
					if window, exists := s.Windows[winId]; exists {
						window.Dirty = true
					}
					s.windowsMu.RUnlock()
				}
			}
		}

		// Track on screen for the active grid
		s.topLine = topLine
		s.botLine = botLine
		s.curLine = curLine
		s.lineCount = lineCount

		if f, ok := viewportArgs[7].(float64); ok {
			s.scrollDelta = f
		} else {
			s.scrollDelta = float64(utils.ReflectToInt(viewportArgs[7]))
		}

		s.emitEvent("viewport_changed", map[string]interface{}{
			"top_line":   topLine,
			"bot_line":   botLine,
			"cur_line":   curLine,
			"line_count": lineCount,
		})
	}
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
		grid = NewGrid(0, 0)
		grid.ID = gridId
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
					Classes:   grid.Cells[i][j].Classes,
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
			window.Dirty = true

			// If this is a small 1x1 window, it's probably not a completion window
			if width == 1 && height == 1 {
				window.IsPopupmenu = false
			}

			utils.Log(fmt.Sprintf("Updated window %d dimensions to %dx%d", winId, width, height))
		}
	}
	s.windowsMu.RUnlock()
	grid.DirtyRows = make([]bool, height)
	grid.OptimizedRows = make([][]*Cell, height)
	grid.CachedTokens = make([][]*Token, height)
	for i := range grid.DirtyRows {
		grid.DirtyRows[i] = true
	}
}

func (s *Screen) gridCursorGoto(gridId int, row int, col int) {
	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}

	previousRow := grid.Cursor.Row
	grid.Cursor.Row = row
	grid.Cursor.Col = col

	s.ActiveGrid = gridId
	s.ActiveWindow = s.GridToWindow[gridId]
	s.UpdateCursor()
	utils.Log(fmt.Sprintf("gridCursorGoto row=%d col=%d activeWindowId=%d", row, col, s.ActiveWindow))
	if row >= 0 && row < grid.Height {
		grid.DirtyRows[row] = true
	}
	if previousRow != row && previousRow >= 0 && previousRow < len(grid.DirtyRows) {
		grid.DirtyRows[previousRow] = true
	}
	// Mark the window dirty so render() emits content-updated
	if winId, exists := s.GridToWindow[gridId]; exists {
		s.windowsMu.RLock()
		if window, exists := s.Windows[winId]; exists {
			window.Dirty = true
		}
		s.windowsMu.RUnlock()
	}
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
	s.highlightsMu.Lock()
	s.Highlights[0] = &Highlight{
		Foreground: s.DefaultFg,
		Background: s.DefaultBg,
		Special:    s.DefaultSp,
	}
	s.highlightsMu.Unlock()

	s.markAllWindowsDirty()
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
			highlight.HasForeground = true
		}

		if bg, ok := rgbAttrs["background"]; ok {
			highlight.Background = utils.ReflectToInt(bg)
			highlight.HasBackground = true
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

		addForegroundColorClass(highlight.fgHex())
		addBackgroundColorClass(highlight.bgHex())

		hlClasses := highlight.getClasses()
		hlClassesStr := mapClassesString(hlClasses)
		var effectiveHlId int
		var existsHlId bool
		effectiveHlIdsMu.Lock()
		if effectiveHlId, existsHlId = effectiveHlIds[hlClassesStr]; !existsHlId {
			effectiveHlIds[hlClassesStr] = id
			effectiveHlId = id
		}
		effectiveHlIdsMu.Unlock()
		addIdClasses(effectiveHlId, hlClasses)
	}

	emitHighlightCSS(s.ctx)
	s.markAllWindowsDirty()
	s.scheduleRender()
}

func (s *Screen) modeChange(mode string) {
	s.Mode = mode
	s.emitEvent("mode-changed", mode)
}

func (s *Screen) winHide(args []interface{}) {
	utils.Log("winHide args", args)
	gridId := utils.ReflectToInt(args[0])
	winId := s.GridToWindow[gridId]
	if winId == 0 {
		return
	}
	s.windowsMu.Lock()
	window, exists := s.Windows[winId]
	if exists {
		window.Hidden = true
	}
	s.windowsMu.Unlock()
	utils.Log(fmt.Sprintf("winHide hiding window with id=%d, s.Windows:", winId), s.Windows)
	if exists && osWindowMgr != nil && osWindowMgr.IsFloatWindow(winId) {
		osWindowMgr.CloseFloatWindow(winId)
	}
	s.emitEvent("hide-window", winId)
}

func (s *Screen) closeTrek(windowIds []int) {
	s.windowsMu.Lock()
	for _, winId := range windowIds {
		_, exists := s.Windows[winId]
		if exists {
			delete(s.Windows, winId)
		}
		for grid, gridWinId := range s.GridToWindow {
			if gridWinId == winId {
				delete(s.GridToWindow, grid)
			}
		}
	}
	s.windowsMu.Unlock()
}

func (s *Screen) winClose(args []interface{}) {
	if len(args) < 1 {
		return
	}

	gridId := utils.ReflectToInt(args[0])
	utils.Log(fmt.Sprintf("winClose closing gridId:%d", gridId))
	s.windowsMu.Lock()
	for grid, winId := range s.GridToWindow {
		if grid == gridId {
			win, exists := s.Windows[winId]
			if exists && win.IsFloating() {
				if osWindowMgr != nil && osWindowMgr.IsFloatWindow(winId) {
					osWindowMgr.CloseFloatWindow(winId)
				} else if win.ZIndex == 69420 {
					s.emitEvent("preview-window-closed", winId)
				} else {
					s.emitEvent("floating_window_closed", winId)
				}
			}
			// Close OS window for all non-floating windows
			if exists && !win.IsFloating() && osWindowMgr != nil {
				osWindowMgr.CloseWindow(winId)
			}
			delete(s.GridToWindow, grid)
			delete(s.Windows, winId)
			utils.Log(fmt.Sprintf("winClose Window %d closed (grid %d)", winId, gridId))
			utils.Log(fmt.Sprintf("winClose got %d windows now", len(s.Windows)))
			break
		}
	}
	s.windowsMu.Unlock()
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

	s.emitEvent("cmdline_show", map[string]interface{}{
		"content": content,
		"pos":     pos,
		"firstc":  firstc,
		"prompt":  prompt,
		"indent":  indent,
	})
}

func (s *Screen) winPos(args []interface{}) {
	if len(args) < 6 {
		return
	}

	gridId := utils.ReflectToInt(args[0])
	nwindow := args[1].(nvim.Window)
	winId, _ := strconv.Atoi(strings.Split(nwindow.String(), ":")[1])
	width := utils.ReflectToInt(args[4])
	height := utils.ReflectToInt(args[5])

	s.windowsMu.RLock()
	window, exists := s.Windows[winId]
	if !exists {
		window = NewWindow(winId, s.Grids[gridId])
		s.Windows[winId] = window
		s.GridToWindow[gridId] = winId
		s.emitEvent("window_opened", struct{}{})
		utils.Log(fmt.Sprintf("winPos spawned with id=%d got %d windows now", winId, len(s.Windows)))
	}
	if winId != 0 {
		buffer, error := getWindowBuffer(winId)
		if error == nil && buffer != nil {
			window.Buffer = buffer
		}
	}
	window.Width = width
	window.Height = height
	window.Hidden = false
	window.Dirty = true
	// Line numbers are rendered by the frontend — neovim's gutter is disabled
	window.lineNumbers = true
	window.relativeLineNumbers = true
	utils.Log(fmt.Sprintf("winPos id=%d gridId=%d Width=%d Height=%d", winId, gridId, width, height))
	s.windowsMu.RUnlock()

	// Ensure grid exists
	if _, exists := s.Grids[gridId]; !exists {
		s.gridResize(gridId, width, height)
	}

	// Every normal window gets its own OS window
	if osWindowMgr != nil {
		osWindowMgr.CreateWindow(winId, gridId)
	}
}

func (s *Screen) winFloatPos(args []interface{}) {
	utils.Log("winFloatPos: ", args)
	if len(args) < 8 {
		utils.Log(fmt.Sprintf("winFloatPos not enough args %d", len(args)))
		return
	}

	gridId := utils.ReflectToInt(args[0])
	str := strings.TrimSpace(s.Grids[gridId].toString())
	//NOTE: seems arbitrary, but theres some weird floating windows with bs content
	// idk wtf they are and i dont care. i want them gone. cant imagine they could be important with 3 chars
	if len(str) < 4 {
		return
	}
	utils.Log(fmt.Sprintf("winFloatPos computing args for gridId %d", gridId), args)
	nwindow := args[1].(nvim.Window)
	winId, _ := strconv.Atoi(strings.Split(nwindow.String(), ":")[1])
	anchor := args[2].(string)
	anchorGrid := utils.ReflectToInt(args[3])
	anchorRow := utils.ReflectToFloat(args[4])
	anchorCol := utils.ReflectToFloat(args[5])
	focusable := args[6].(bool)
	zIndex := utils.ReflectToInt(args[7])
	utils.Log(fmt.Sprintf("winFloatPos winId: %d anchorGrid: %d row: %f col %f", winId, anchorGrid, anchorRow, anchorCol), args)

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

	if winId != 0 {
		buffer, error := getWindowBuffer(winId)
		if error == nil && buffer != nil {
			window.Buffer = buffer
		} else {
			utils.Log(fmt.Sprintf("winFloatPos could not get the filetype for window with id=%d error:%s", winId, error.Error()))
		}
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
	window.Dirty = true

	// Check if this might be a completion window
	// Completion windows are typically floating windows with specific characteristics
	if window.Width > 1 && window.Height > 1 {
		// This is a heuristic - we might need to refine this
		window.IsPopupmenu = true
	}

	// Update grid to window mapping
	s.GridToWindow[gridId] = winId

	// Floating windows anchored to the global grid should be separate OS windows
	if anchorGrid == 1 && osWindowMgr != nil {
		osWindowMgr.CreateFloatWindow(winId, gridId, window.Width, window.Height)
	}

	utils.Log(fmt.Sprintf("Floating window %d anchored at grid %d (%f,%f) with z-index %d",
		winId, anchorGrid, anchorRow, anchorCol, zIndex))
}

func (s *Screen) winExternalPos(args []interface{}) {
	if len(args) < 2 {
		return
	}

	gridId := utils.ReflectToInt(args[0])
	nwindow := args[1].(nvim.Window)
	winId, _ := strconv.Atoi(strings.Split(nwindow.String(), ":")[1])

	utils.Log(fmt.Sprintf("winExternalPos gridId=%d winId=%d", gridId, winId))

	s.windowsMu.Lock()
	window, exists := s.Windows[winId]
	if !exists {
		window = NewWindow(winId, s.Grids[gridId])
		s.Windows[winId] = window
		s.GridToWindow[gridId] = winId
	}
	window.Hidden = false
	window.Dirty = true
	s.windowsMu.Unlock()

	if _, exists := s.Grids[gridId]; !exists {
		s.gridResize(gridId, 80, 24)
	}

	if winId != 0 {
		buffer, err := getWindowBuffer(winId)
		if err == nil && buffer != nil {
			window.Buffer = buffer
		}
	}

	if osWindowMgr != nil {
		osWindowMgr.CreateWindow(winId, gridId)
	}
}

func (s *Screen) markAllWindowsDirty() {
	s.windowsMu.RLock()
	for _, window := range s.Windows {
		window.Dirty = true
	}
	s.windowsMu.RUnlock()
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

	for winId, window := range s.Windows {
		if window.Hidden {
			continue
		}
		if !window.Dirty {
			continue
		}
		window.Dirty = false

		grid := window.Grid
		isOSFloat := osWindowMgr != nil && osWindowMgr.IsFloatWindow(winId)

		if !window.IsFloating() || window.ZIndex == 69420 || isOSFloat {
			filetype := ""
			bufNr := 0
			if window.Buffer != nil {
				filetype = (*window.Buffer).Filetype
				bufNr = (*window.Buffer).BufNr
			}
			if filetype == "markdown" {
				utils.Log(fmt.Sprintf("render: winId=%d filetype=%s bufNr=%d", winId, filetype, bufNr))
			}
			cursorLine := -1
			if window.Cursor != nil {
				cursorLine = window.Cursor.Row - 1 // convert 1-indexed to 0-indexed
			}
			s.emitEvent("content-updated", map[string]interface{}{
				"winId":          winId,
				"updatedContent": s.optimizeGrid(grid, filetype, bufNr, cursorLine),
			})
		} else {
			// Skip rendering blink-cmp floating windows (handled by native completion menu)
			if window.Buffer != nil && (*window.Buffer).Filetype == "blink-cmp-menu" {
				continue
			}
			if s.app != nil {
				s.app.Event.Emit("content-updated", map[string]interface{}{
					"winId":          winId,
					"floating":       true,
					"updatedContent": s.renderFloatingWindow(window),
				})
			}
		}
	}
}

// handleCmdlinePos processes the cmdline_pos event
func (s *Screen) handleCmdlinePos(args []interface{}) {
	cmdlineArgs := args[0].([]interface{})
	if len(cmdlineArgs) < 2 {
		return
	}

	pos := utils.ReflectToInt(cmdlineArgs[0])
	level := utils.ReflectToInt(cmdlineArgs[1])

	s.emitEvent("cmdline_pos", map[string]interface{}{
		"pos":   pos,
		"level": level,
	})
}

func (s *Screen) EmitFloatingWindows() {
	s.windowsMu.RLock()
	defer s.windowsMu.RUnlock()

	// Only emit if any floating window is dirty (newly created/updated)
	hasNewFloating := false
	for _, window := range s.Windows {
		if window.IsFloating() && window.ZIndex != 69420 && window.Dirty {
			hasNewFloating = true
			break
		}
	}
	if !hasNewFloating {
		return
	}

	var floatingWindows []map[string]interface{}

	for winId, window := range s.Windows {
		if !window.IsFloating() {
			continue
		}
		// Skip blink-cmp windows -- handled by native completion menu
		if window.Buffer != nil && (*window.Buffer).Filetype == "blink-cmp-menu" {
			continue
		}
		// Skip OS float windows -- rendered in their own OS window
		if osWindowMgr != nil && osWindowMgr.IsFloatWindow(winId) {
			continue
		}
		if window.ZIndex == 69420 {
			if window.Dirty {
				previewWindow := s.mapWindowInfo(window, winId)
				s.emitEvent("preview-window", previewWindow)
			}
			continue
		}
		// Get the associated grid
		_, exists := s.Grids[window.Grid.ID]
		if !exists {
			continue
		}
		windowInfo := s.mapWindowInfo(window, winId)
		floatingWindows = append(floatingWindows, windowInfo)
	}

	if len(floatingWindows) > 0 {
		s.emitEvent("floating_windows", floatingWindows)
	}
}

// EmitCurrentState re-emits the current state to all connected clients
// This is useful for late-connecting clients (like Playwright tests) that miss the initial events
func (s *Screen) EmitCurrentState() {
	emitHighlightCSS(s.ctx)

	// Re-emit content for all windows
	s.windowsMu.RLock()
	defer s.windowsMu.RUnlock()
	for winId, window := range s.Windows {
		if window.Hidden {
			continue
		}
		grid, exists := s.Grids[window.Grid.ID]
		if !exists {
			continue
		}
		filetype := ""
		bufNr := 0
		if window.Buffer != nil {
			filetype = (*window.Buffer).Filetype
			bufNr = (*window.Buffer).BufNr
		}
		cursorLine := -1
		if window.Cursor != nil {
			cursorLine = window.Cursor.Row - 1
		}
		if s.app != nil {
			s.app.Event.Emit("content-updated", map[string]interface{}{
				"winId":          winId,
				"updatedContent": s.optimizeGrid(grid, filetype, bufNr, cursorLine),
			})
		}
	}
}

func (s *Screen) mapWindowInfo(window *Window, winId int) map[string]interface{} {
	anchorWindow, ok := s.GridToWindow[window.AnchorGrid]
	if !ok || anchorWindow == 0 {
		// Floating windows anchored to the global grid (1) or an unknown grid
		// should display on the active window
		anchorWindow = s.ActiveWindow
	}
	utils.Log(fmt.Sprintf("mapWindowInfo winId=%d anchorGrid=%d anchorWindow=%d", winId, window.AnchorGrid, anchorWindow))
	windowInfo := map[string]interface{}{
		"id":           winId,
		"gridId":       window.Grid.ID,
		"anchorWindow": anchorWindow,
		"anchor":       window.Anchor,
		"col":          window.StartCol,
		"row":          window.StartRow,
		"width":        window.Width,
		"height":       window.Height,
		"zIndex":       window.ZIndex,
		"focusable":    window.Focusable,
		"isPopup":      window.IsPopupmenu,
	}
	if window.Buffer != nil {
		windowInfo["filetype"] = &window.Buffer.Filetype
	}
	//NOTE: need hex encoding for some nerdfont stuff (e.g. completion window)
	windowInfo["isHex"] = isHex(window)
	return windowInfo
}

func (s *Screen) setTableMetadata(bufNr int, tables []TableMeta) {
	s.tableMetadataMu.Lock()
	defer s.tableMetadataMu.Unlock()
	s.tableMetadata[bufNr] = tables
}

func (s *Screen) getTableMetaForLine(bufNr int, bufferLine int) *TableMeta {
	if bufNr == 0 {
		return nil
	}
	s.tableMetadataMu.RLock()
	defer s.tableMetadataMu.RUnlock()
	tables, exists := s.tableMetadata[bufNr]
	if !exists {
		return nil
	}
	for i := range tables {
		// Extend range by 1 on each side to capture box-drawing border rows
		// added by render-markdown plugins (e.g., ┌─┬─┐ and └─┴─┘)
		if bufferLine >= tables[i].StartLine-1 && bufferLine < tables[i].EndLine+1 {
			return &tables[i]
		}
	}
	return nil
}

func parseMarkdownTables(tablesRaw []interface{}) []TableMeta {
	var tables []TableMeta
	for _, raw := range tablesRaw {
		tableData, ok := raw.([]interface{})
		if !ok || len(tableData) < 3 {
			continue
		}
		startLine := utils.ReflectToInt(tableData[0])
		endLine := utils.ReflectToInt(tableData[1])
		alignmentsRaw, ok := tableData[2].([]interface{})
		if !ok {
			continue
		}
		alignments := make([]string, len(alignmentsRaw))
		for i, a := range alignmentsRaw {
			if s, ok := a.(string); ok {
				alignments[i] = s
			} else {
				alignments[i] = "left"
			}
		}
		tables = append(tables, TableMeta{
			StartLine:  startLine,
			EndLine:    endLine,
			Alignments: alignments,
		})
	}
	return tables
}

func (s *Screen) setImageMetadata(bufNr int, images []ImageMeta) {
	s.imageMetadataMu.Lock()
	defer s.imageMetadataMu.Unlock()
	s.imageMetadata[bufNr] = images
}

func (s *Screen) getImageMetaForLine(bufNr int, bufferLine int) *ImageMeta {
	if bufNr == 0 {
		return nil
	}
	s.imageMetadataMu.RLock()
	defer s.imageMetadataMu.RUnlock()
	images, exists := s.imageMetadata[bufNr]
	if !exists {
		return nil
	}
	for i := range images {
		if images[i].Line == bufferLine {
			return &images[i]
		}
	}
	return nil
}

func parseMarkdownImages(imagesRaw []interface{}) []ImageMeta {
	var images []ImageMeta
	for _, raw := range imagesRaw {
		imageData, ok := raw.([]interface{})
		if !ok || len(imageData) < 3 {
			continue
		}
		line := utils.ReflectToInt(imageData[0])
		url, ok := imageData[1].(string)
		if !ok {
			continue
		}
		alt, _ := imageData[2].(string)

		// Rewrite local paths to use the local-image server
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			encoded := base64.URLEncoding.EncodeToString([]byte(url))
			url = "/local-image/" + encoded
		}

		images = append(images, ImageMeta{
			Line:    line,
			URL:     url,
			AltText: alt,
		})
	}
	return images
}

func (s *Screen) setHeadingMetadata(bufNr int, headings map[int]int) {
	s.headingMetadataMu.Lock()
	defer s.headingMetadataMu.Unlock()
	s.headingMetadata[bufNr] = headings
}

func (s *Screen) getHeadingLevel(bufNr int, bufferLine int) int {
	if bufNr == 0 {
		return 0
	}
	s.headingMetadataMu.RLock()
	defer s.headingMetadataMu.RUnlock()
	headings, exists := s.headingMetadata[bufNr]
	if !exists {
		return 0
	}
	return headings[bufferLine]
}

func parseMarkdownHeadings(headingsRaw []interface{}) map[int]int {
	headings := make(map[int]int)
	for _, raw := range headingsRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		line := utils.ReflectToInt(entry[0])
		level := utils.ReflectToInt(entry[1])
		headings[line] = level
	}
	return headings
}

func (s *Screen) setTaskMetadata(bufNr int, tasks map[int]bool) {
	s.taskMetadataMu.Lock()
	defer s.taskMetadataMu.Unlock()
	s.taskMetadata[bufNr] = tasks
}

func (s *Screen) getTaskMetaForLine(bufNr int, bufferLine int) (checked bool, isTask bool) {
	if bufNr == 0 {
		return false, false
	}
	s.taskMetadataMu.RLock()
	defer s.taskMetadataMu.RUnlock()
	tasks, exists := s.taskMetadata[bufNr]
	if !exists {
		return false, false
	}
	checked, isTask = tasks[bufferLine]
	return checked, isTask
}

func parseMarkdownTasks(tasksRaw []interface{}) map[int]bool {
	tasks := make(map[int]bool)
	for _, raw := range tasksRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		line := utils.ReflectToInt(entry[0])
		checked := utils.ReflectToInt(entry[1]) == 1
		tasks[line] = checked
	}
	return tasks
}

func (s *Screen) setCodeBlockMetadata(bufNr int, blocks []CodeBlockMeta) {
	s.codeBlockMetadataMu.Lock()
	defer s.codeBlockMetadataMu.Unlock()
	s.codeBlockMetadata[bufNr] = blocks
}

func (s *Screen) getCodeBlockMetaForLine(bufNr int, bufferLine int) *CodeBlockMeta {
	if bufNr == 0 {
		return nil
	}
	s.codeBlockMetadataMu.RLock()
	defer s.codeBlockMetadataMu.RUnlock()
	blocks, exists := s.codeBlockMetadata[bufNr]
	if !exists {
		return nil
	}
	for i := range blocks {
		if bufferLine >= blocks[i].StartLine && bufferLine < blocks[i].EndLine {
			return &blocks[i]
		}
	}
	return nil
}

func parseMarkdownCodeBlocks(blocksRaw []interface{}) []CodeBlockMeta {
	var blocks []CodeBlockMeta
	for _, raw := range blocksRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		startLine := utils.ReflectToInt(entry[0])
		endLine := utils.ReflectToInt(entry[1])
		blocks = append(blocks, CodeBlockMeta{
			StartLine: startLine,
			EndLine:   endLine,
		})
	}
	return blocks
}

func (s *Screen) setInlineCodeMetadata(bufNr int, codes map[int][]ColRange) {
	s.inlineCodeMetadataMu.Lock()
	defer s.inlineCodeMetadataMu.Unlock()
	s.inlineCodeMetadata[bufNr] = codes
}

func (s *Screen) getInlineCodeRanges(bufNr int, bufferLine int) []ColRange {
	if bufNr == 0 {
		return nil
	}
	s.inlineCodeMetadataMu.RLock()
	defer s.inlineCodeMetadataMu.RUnlock()
	lines, exists := s.inlineCodeMetadata[bufNr]
	if !exists {
		return nil
	}
	return lines[bufferLine]
}

func parseMarkdownInlineCode(codesRaw []interface{}) map[int][]ColRange {
	codes := make(map[int][]ColRange)
	for _, raw := range codesRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 3 {
			continue
		}
		line := utils.ReflectToInt(entry[0])
		startCol := utils.ReflectToInt(entry[1])
		endCol := utils.ReflectToInt(entry[2])
		codes[line] = append(codes[line], ColRange{StartCol: startCol, EndCol: endCol})
	}
	return codes
}

func (s *Screen) GetActiveWindow() *Window {
	return s.Windows[s.ActiveWindow]
}
