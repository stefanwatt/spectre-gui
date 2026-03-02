package main

import (
	"context"
	"encoding/base64"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"nvim-gui/neovim"
	"nvim-gui/utils"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var ctx context.Context

type App struct {
	ctx              context.Context
	File             string
}

func NewApp() *App {
	return &App{
	}
}

func (a *App) mounted(ctx context.Context) {
	// Force WebKit to recalculate its viewport with the compositor-allocated size.
	// On Wayland (e.g. Sway), the WM assigns the final window dimensions asynchronously,
	// so WebKit may initialize with the GTK default size. Re-setting the size after DOM
	// ready triggers gtk_window_resize and causes WebKit to update its viewport.
	w, h := Runtime.WindowGetSize(ctx)
	Runtime.WindowSetSize(ctx, w, h)
	go neovim.StartListening(a.ctx)
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	utils.SetupLog()
}

// ReadLocalImage decodes a base64-encoded local file path and returns it as a data URL.
// This allows the frontend to load local images in both dev and production modes.
func (a *App) ReadLocalImage(encoded string) string {
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	absPath := string(decoded)
	ext := strings.ToLower(filepath.Ext(absPath))
	if !imageExtensions[ext] {
		return ""
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return ""
	}
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)
}
