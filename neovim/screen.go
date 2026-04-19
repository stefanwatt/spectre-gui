package neovim

import (
	"context"
	"nvim-gui/rendering"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Screen struct {
	Height int // in number of cells
	Width  int // in number of cells

	// Windows and ActiveWindow are retained for GetActiveWindow(),
	// which is consumed by the picker features.
	Windows      map[int]*Window
	ActiveWindow int
	windowsMu    sync.RWMutex

	// Markdown metadata — populated by RPC handlers, consumed by rendering pipeline.
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

func NewScreen(ctx context.Context, cols, rows int, app *application.App) *Screen {
	return &Screen{
		Width:  cols,
		Height: rows,

		Windows: make(map[int]*Window),

		tableMetadata:      make(map[int][]rendering.TableMeta),
		imageMetadata:      make(map[int][]rendering.ImageMeta),
		headingMetadata:    make(map[int]map[int]int),
		taskMetadata:       make(map[int]map[int]bool),
		codeBlockMetadata:  make(map[int][]rendering.CodeBlockMeta),
		inlineCodeMetadata: make(map[int]map[int][]rendering.ColRange),
	}
}

func (s *Screen) Resize(width, height int) {
	s.Width = width
	s.Height = height
	NvimClient.TryResizeUI(width, height)
}

func (s *Screen) SetFileExplorerPreviewSizePixels(widthPx, heightPx int) {
	// cols, rows := PreviewGridSizeFromPixels(widthPx, heightPx)
	// currentCols := cols / 2
	// if currentCols < 1 {
	// 	currentCols = 1
	// }
	// currentRows := rows
	// if currentRows < 1 {
	// 	currentRows = 1
	// }
	// if err := SetMiniFilesWindowOverrides(currentCols, cols, currentRows, rows); err != nil {
	// 	log.Debug(fmt.Sprintf("[minifiles] failed to set window overrides current=%dx%d preview=%dx%d err=%v",
	// 		currentCols, currentRows, cols, rows, err))
	// }
}

func (s *Screen) GetActiveWindow() *Window {
	return s.Windows[s.ActiveWindow]
}
