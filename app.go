package main

import (
	"context"
	"net/url"
	appruntime "nvim-gui/app/runtime"
	"nvim-gui/core/projection"
	fileexplorer "nvim-gui/features/file-explorer"
	"nvim-gui/neovim"
	"nvim-gui/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	ctx          context.Context
	App          *application.App
	runtime      *appruntime.Runtime
	fileExplorer *fileexplorer.FileExplorer
}

type wailsUIEmitter struct {
	app *application.App
}

func (w wailsUIEmitter) Emit(name string, payload any) {
	if w.app == nil {
		return
	}
	w.app.Event.Emit(name, payload)
}

func NewApp() *App {
	return &App{}
}

func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.ctx = ctx
	utils.SetupLog()
	fileExplorer := fileexplorer.NewFileExplorer(&neovim.NvimAdapter{})
	a.fileExplorer = fileExplorer
	neovim.SetFileExplorer(fileExplorer)
	a.runtime = appruntime.New(nil, fileExplorer)
	emitter := wailsUIEmitter{app: a.App}
	a.runtime.SetEmitter(emitter)
	fileExplorer.SetEmitter(emitter)
	a.runtime.SetProjector(projection.CompositeProjector{
		Projectors: []projection.Projector{
			projection.NewLayoutContentProjector(),
			projection.EventsProjector{},
			projection.NewFileExplorerProjector(fileExplorer),
		},
	})
	a.runtime.Start(ctx)
	neovim.SetEventSink(a.runtime)

	go neovim.StartListening(ctx)

	return nil
}

func (a *App) OnResize(width, height int) {
	rows, cols := neovim.CalculateGridSize(width, height)
	if neovim.NvimScreen != nil {
		neovim.NvimScreen.Resize(cols, rows)
	}
}

func (a *App) OnFileExplorerPreviewResize(width, height int) {
	if a.fileExplorer != nil {
		a.fileExplorer.ResizePreviewPixels(width, height)
	}
}

func (a *App) OnFileExplorerConfirmChoice(choice int) {
	if a.fileExplorer == nil {
		return
	}
	a.fileExplorer.HandleConfirmChoice(choice)
}

// RequestState triggers emission of current state to frontend
func (a *App) RequestState() {
	if neovim.NvimScreen != nil {
		// TODO: delete after a while if we dont notice problem with it being gone
		// neovim.NvimScreen.EmitCurrentState()
	}
}

// ReadLocalImage returns an asset-server URL for a local image path.
func (a *App) ReadLocalImage(imagePath string) string {
	return localImageURL(imagePath)
}

func normalizeLocalImagePath(imagePath string) string {
	if strings.HasPrefix(imagePath, "file://") {
		parsed, err := url.Parse(imagePath)
		if err == nil {
			if unescaped, err := url.PathUnescape(parsed.Path); err == nil {
				imagePath = unescaped
			} else {
				imagePath = parsed.Path
			}
		}
	}
	if imagePath == "~" || strings.HasPrefix(imagePath, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if imagePath == "~" {
				return home
			}
			return filepath.Join(home, strings.TrimPrefix(imagePath, "~/"))
		}
	}
	return imagePath
}
