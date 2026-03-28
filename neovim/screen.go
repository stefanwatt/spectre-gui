package neovim

import (
	"context"
	"fmt"
	"math"
	"nvim-gui/rendering"
	"nvim-gui/utils"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/neovim/go-client/nvim"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Screen struct {
	margins           []int
	Height            int // in number of cells
	Width             int // in number of cells
	topLine           int
	botLine           int
	curLine           int
	lineCount         int
	scrollDelta       float64
	ctx               context.Context
	Grids             map[int]*Grid
	Windows           map[int]*Window // Map of window IDs to Window objects
	GridToWindow      map[int]int     // Map of grid IDs to window IDs
	ActiveWindow      int
	Mode              string
	windowsMu         sync.RWMutex // Mutex for Windows map
	layout            *rendering.GridLayout
	ColorColumns      []int  // columns where colorcolumn should render (e.g. [80])
	ColorColumnColor  string // hex color e.g. "#2a2a3a"
	CursorLineEnabled bool
	CursorLineColor   string // hex color e.g. "#2a2a3a"
	FileExplorer      *FileExplorer

	tableMetadata        map[int][]rendering.TableMeta
	tableMetadataMu      sync.RWMutex
	imageMetadata        map[int][]rendering.ImageMeta
	imageMetadataMu      sync.RWMutex
	headingMetadata      map[int]map[int]int
	headingMetadataMu    sync.RWMutex
	taskMetadata         map[int]map[int]bool
	taskMetadataMu       sync.RWMutex
	codeBlockMetadata    map[int][]rendering.CodeBlockMeta
	codeBlockMetadataMu  sync.RWMutex
	inlineCodeMetadata   map[int]map[int][]rendering.ColRange
	inlineCodeMetadataMu sync.RWMutex
}

func (s *Screen) syncFileExplorerMode() {
	if s.FileExplorer == nil {
		return
	}
	s.FileExplorer.CurrentWinMode = s.Mode
}

func NewScreen(ctx context.Context, cols, rows int, app *application.App) *Screen {
	// Create grids map and add default grid
	grids := make(map[int]*Grid)
	grids[1] = NewGrid(rows, cols)

	// Create default highlight

	return &Screen{
		ctx:          ctx,
		Width:        cols,
		Height:       rows,
		Grids:        grids,
		Windows:      make(map[int]*Window),
		GridToWindow: make(map[int]int),
		Mode:         "normal",
		margins:      make([]int, 4), // Initialize margins slice
		layout:       rendering.NewGridLayout(),

		tableMetadata:      make(map[int][]rendering.TableMeta),
		imageMetadata:      make(map[int][]rendering.ImageMeta),
		headingMetadata:    make(map[int]map[int]int),
		taskMetadata:       make(map[int]map[int]bool),
		codeBlockMetadata:  make(map[int][]rendering.CodeBlockMeta),
		inlineCodeMetadata: make(map[int]map[int][]rendering.ColRange),
	}
}

func (s *Screen) Resize(width, height int) {
	NvimScreen.Width = width
	NvimScreen.Height = height
	NvimClient.TryResizeUI(width, height)
	s.applyFileExplorerPreviewSize()
}

func (s *Screen) SetFileExplorerPreviewSizePixels(widthPx, heightPx int) {
	cols, rows := PreviewGridSizeFromPixels(widthPx, heightPx)
	currentCols := cols / 2
	if currentCols < 1 {
		currentCols = 1
	}
	currentRows := rows
	if currentRows < 1 {
		currentRows = 1
	}
	if err := SetMiniFilesWindowOverrides(currentCols, cols, currentRows, rows); err != nil {
		log.Debug(fmt.Sprintf("[minifiles] failed to set window overrides current=%dx%d preview=%dx%d err=%v",
			currentCols, currentRows, cols, rows, err))
	}
	if s.FileExplorer != nil {
		if s.FileExplorer.PreviewTargetCols == cols && s.FileExplorer.PreviewTargetRows == rows {
			return
		}
		s.FileExplorer.SetPreviewTargetSize(cols, rows)
	}
	s.applyFileExplorerPreviewSize()
}

func (s *Screen) applyFileExplorerPreviewSize() {
	if s.FileExplorer == nil || !s.FileExplorer.HasPreviewTargetSize() {
		return
	}
	previewWinId := s.FileExplorer.Preview.GetWinId()
	if previewWinId < 1 {
		return
	}
	err := resizeFloatingWindow(previewWinId, s.FileExplorer.PreviewTargetCols, s.FileExplorer.PreviewTargetRows)
	if err != nil {
		log.Debug(fmt.Sprintf("[minifiles] failed to resize preview winId=%d cols=%d rows=%d err=%v",
			previewWinId, s.FileExplorer.PreviewTargetCols, s.FileExplorer.PreviewTargetRows, err))
	}
}

func (s *Screen) GridScroll(gridId, top, bot, left, right, rows, cols int) {
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

func (s *Screen) GridLine(gridId, row, col int, cells []interface{}) {
	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}
	if row >= grid.Height || col >= grid.Width {
		log.Debug(fmt.Sprintf("Row %d or col %d out of bounds for grid %d (max: %d,%d)",
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

				highlightsMu.RLock()
				highlight, exists := highlights[hl]
				if !exists || highlight == nil {
					highlight = highlights[0]
					hl = 0
				}
				highlightsMu.RUnlock()

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
			}
		}
	}

	grid.DirtyRows[row] = true
}

func (s *Screen) GridClear(gridId int) {
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

func (s *Screen) HandleWinViewport(args []interface{}) {
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

		EmitEvent("viewport_changed", map[string]interface{}{
			"top_line":   topLine,
			"bot_line":   botLine,
			"cur_line":   curLine,
			"line_count": lineCount,
		})
	}
}

func (s *Screen) HandleWinViewportMargins(args []interface{}) {
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

func (s *Screen) GridResize(gridId, width, height int) {
	if gridId > 2 {
		log.Debug(fmt.Sprintf("gridResize gridId: %d, width: %d, height: %d", gridId, width, height))
	}
	grid, exists := s.Grids[gridId]
	if !exists {
		grid = NewGrid(0, 0)
		grid.ID = gridId
		s.Grids[gridId] = grid
	}

	grid.Resize(width, height)

	// Update window dimensions if this grid is associated with a window
	s.windowsMu.RLock()
	if winId, exists := s.GridToWindow[gridId]; exists {
		if window, exists := s.Windows[winId]; exists {
			window.Resize(width, height)
		}
	}
	s.windowsMu.RUnlock()
}

func (s *Screen) GridCursorGoto(gridId, row, col int) {
	grid, exists := s.Grids[gridId]
	if !exists {
		return
	}

	previousRow := grid.Cursor.Row
	grid.Cursor.Row = row
	grid.Cursor.Col = col

	s.ActiveWindow = s.GridToWindow[gridId]
	s.UpdateCursor()
	if s.FileExplorer != nil {
		entry, err := s.FileExplorer.updateCursor(s.ActiveWindow, row, col)
		if err == nil {
			log.Debug(fmt.Sprintf("[fileexplorer] cursor on entry: text=%s isDir=%v", entry.Text, entry.IsDir))
			s.syncFileExplorerMode()
			EmitEvent("file-explorer-update", s.FileExplorer)
		}
	}
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

func (s *Screen) ModeChange(mode string) {
	s.Mode = mode
	EmitEvent("mode-changed", mode)
	if s.FileExplorer != nil {
		s.syncFileExplorerMode()
		EmitEvent("file-explorer-update", s.FileExplorer)
	}
}

func (s *Screen) WinHide(args []interface{}) {
	log.Debug("winHide args", args)
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
	log.Debug(fmt.Sprintf("winHide hiding window with id=%d, s.Windows:", winId), s.Windows)
	EmitEvent("file-explorer-confirm-prompt-hide", struct{}{})
	EmitEvent("hide-window", winId)
	s.updateLayout()
}

func (s *Screen) WinClose(args []interface{}) {
	if len(args) < 1 {
		return
	}

	gridId := utils.ReflectToInt(args[0])
	// Find the grid associated with this window
	log.Debug(fmt.Sprintf("winClose closing gridId:%d", gridId))
	s.windowsMu.RLock()
	closedFileExplorer := false
	for grid, winId := range s.GridToWindow {
		if grid == gridId {
			win, exists := s.Windows[winId]
			if exists && win.IsFloating() {
				if win.IsFileExplorer {
					closedFileExplorer = true
				} else if win.ZIndex == 69420 {
					EmitEvent("preview-window-closed", winId)
				} else {
					EmitEvent("file-explorer-confirm-prompt-hide", struct{}{})
					EmitEvent("floating_window_closed", winId)
				}
			}
			delete(s.GridToWindow, grid)
			delete(s.Windows, winId)
			log.Debug(fmt.Sprintf("winClose Window %d closed (grid %d)", winId, gridId))
			log.Debug(fmt.Sprintf("winClose got %d windows now", len(s.Windows)))
			break
		}
	}
	// Check if all file explorer windows are now gone
	if closedFileExplorer {
		hasFileExplorer := false
		for _, win := range s.Windows {
			if win.IsFileExplorer {
				hasFileExplorer = true
				break
			}
		}
		if !hasFileExplorer {
			EmitEvent("file-explorer-close", struct{}{})
		}
	}
	s.windowsMu.RUnlock()
}

func (s *Screen) HandleCmdlineShow(args []interface{}) {
	arg := args[0].([]interface{})
	cmdline := rendering.RenderCmdline(arg)
	EmitEvent("cmdline_show", cmdline)
}

func (s *Screen) WinPos(args []interface{}) {
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
		EmitEvent("window_opened", struct{}{})
		log.Debug(fmt.Sprintf("winPos spawned with id=%d got %d windows now", winId, len(s.Windows)))
	}
	if winId != 0 {
		buffer, error := GetWindowBuffer(winId)
		if error == nil && buffer != nil {
			window.Buffer = buffer
		}
	}
	window.Width = width
	window.Height = height
	window.StartRow = row
	window.StartCol = col
	window.Hidden = false
	window.Dirty = true
	// Line numbers are rendered by the frontend — neovim's gutter is disabled
	window.lineNumbers = true
	window.relativeLineNumbers = true
	log.Debug(fmt.Sprintf("winPos id=%d gridId=%d StartRow=%d StartCol=%d Width=%d Height=%d", winId, gridId, row, col, width, height))
	s.windowsMu.RUnlock()

	// Handle window positioning
	if _, exists := s.Grids[gridId]; !exists {
		// Create a new grid for this window
		s.GridResize(gridId, width, height)
	}
}

func (s *Screen) updateLayout() {
	prev := s.layout
	s.CalculateGridLayout()
	if prev == nil || !rendering.LayoutEqual(prev, s.layout) {
		EmitEvent("layout-updated", s.layout)
	}
}

func (s *Screen) WinFloatPos(args []interface{}) {
	log.Debug("winFloatPos: ", args)
	if len(args) < 8 {
		log.Debug(fmt.Sprintf("winFloatPos not enough args %d", len(args)))
		return
	}

	gridId := utils.ReflectToInt(args[0])
	str := strings.TrimSpace(s.Grids[gridId].toString())
	// NOTE: seems arbitrary, but theres some weird floating windows with bs content
	// idk wtf they are and i dont care. i want them gone. cant imagine they could be important with 3 chars
	if len(str) < 4 {
		return
	}
	log.Debug(fmt.Sprintf("winFloatPos computing args for gridId %d", gridId), args)
	nwindow := args[1].(nvim.Window)
	winId, _ := strconv.Atoi(strings.Split(nwindow.String(), ":")[1])
	anchor := args[2].(string)
	anchorGrid := utils.ReflectToInt(args[3])
	anchorRow := utils.ReflectToFloat(args[4])
	anchorCol := utils.ReflectToFloat(args[5])
	focusable := args[6].(bool)
	zIndex := utils.ReflectToInt(args[7])
	log.Debug(fmt.Sprintf("winFloatPos winId: %d row: %f col %f", winId, anchorRow, anchorCol), args)

	// Update window tracking
	s.windowsMu.Lock()
	defer s.windowsMu.Unlock()

	window, exists := s.Windows[winId]

	if !exists {
		log.Debug(fmt.Sprintf("winFloatPos adding winId %d", winId))
		window = NewWindow(winId, s.Grids[gridId])
		s.Windows[winId] = window
	} else {
		log.Debug(fmt.Sprintf("winFloatPos already had winId %d", winId))
	}

	existingGrid := s.Grids[gridId]
	if existingGrid != nil {
		window.Width = existingGrid.Width
		window.Height = existingGrid.Height
	}

	if winId != 0 {
		log.Debug(fmt.Sprintf("[minifiles] winFloatPos: calling GetWindowBuffer for winId=%d", winId))
		buffer, error := GetWindowBuffer(winId)
		if error == nil && buffer != nil {
			window.Buffer = buffer
			log.Debug(fmt.Sprintf("[minifiles] winFloatPos: winId=%d filetype=%s bufNr=%d", winId, buffer.Filetype, buffer.BufNr))
			if buffer.Filetype == "minifiles" {
				window.IsFileExplorer = true
				log.Debug(fmt.Sprintf("[minifiles] winFloatPos: marked winId=%d as file explorer", winId))
			}
		} else {
			log.Debug(fmt.Sprintf("winFloatPos could not get the filetype for window with id=%d error:%s", winId, error.Error()))
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

	log.Debug(fmt.Sprintf("Floating window %d anchored at grid %d (%f,%f) with z-index %d",
		winId, anchorGrid, anchorRow, anchorCol, zIndex))
}

func (s *Screen) MarkAllWindowsDirty() {
	s.windowsMu.RLock()
	for _, window := range s.Windows {
		window.Dirty = true
	}
	s.windowsMu.RUnlock()
}

// handleCmdlinePos processes the cmdline_pos event
func (s *Screen) HandleCmdlinePos(args []interface{}) {
	cmdlineArgs := args[0].([]interface{})
	if len(cmdlineArgs) < 2 {
		return
	}

	pos := utils.ReflectToInt(cmdlineArgs[0])
	level := utils.ReflectToInt(cmdlineArgs[1])

	EmitEvent("cmdline_pos", map[string]interface{}{
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
		if window.IsFloating() && window.ZIndex != 69420 && !window.IsFileExplorer && window.Dirty {
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
		if s.emitFileExplorerConfirmPrompt(window) {
			continue
		}
		// Skip blink-cmp windows -- handled by native completion menu
		if window.Buffer != nil && (*window.Buffer).Filetype == "blink-cmp-menu" {
			continue
		}
		// Skip mini.files directory windows -- rendered by custom file explorer UI
		if window.IsFileExplorer {
			continue
		}
		if window.ZIndex == 69420 {
			if window.Dirty {
				previewWindow := s.mapWindowInfo(window, winId)
				EmitEvent("preview-window", previewWindow)
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
		EmitEvent("floating_windows", floatingWindows)
	}
}

// EmitCurrentState re-emits the current state to all connected clients
// This is useful for late-connecting clients (like Playwright tests) that miss the initial events
func (s *Screen) EmitCurrentState() {
	css := rendering.BuildHighlightCSS()
	EmitEvent("highlight-css", css)

	// Force emit layout regardless of dirty flag (late-connecting clients missed the initial emit)
	s.CalculateGridLayout()
	s.layout.ActiveWindowId = s.ActiveWindow
	EmitEvent("layout-updated", s.layout)
	EmitEvent("file-explorer-confirm-prompt-hide", struct{}{})
	if s.FileExplorer != nil {
		s.syncFileExplorerMode()
		EmitEvent("file-explorer-update", s.FileExplorer)
	}

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
		payload := rendering.BuildContentPayload(rendering.ContentInput{
			WindowID:   winId,
			Filetype:   filetype,
			BufNr:      bufNr,
			CursorLine: cursorLine,
			Grid:       s.toRenderingGridData(grid),
			Meta:       nil,
		})
		EmitEvent("content-updated", map[string]interface{}{
			"winId":          payload.WindowID,
			"updatedContent": payload.Content,
		})
	}
}

func (s *Screen) mapWindowInfo(window *Window, winId int) map[string]interface{} {
	anchorWindow, _ := s.GridToWindow[window.AnchorGrid]
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
	// NOTE: need hex encoding for some nerdfont stuff (e.g. completion window)
	windowInfo["isHex"] = isHex(window)
	return windowInfo
}

func (s *Screen) GetActiveWindow() *Window {
	return s.Windows[s.ActiveWindow]
}
