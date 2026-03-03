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

	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	ctx context.Context
	App *application.App
}

func NewApp() *App {
	return &App{}
}

func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.ctx = ctx
	utils.SetupLog()

	// Start neovim after app is ready
	go neovim.StartListening(ctx)

	return nil
}

// OnResize handles window resize events from the frontend
func (a *App) OnResize(width, height int) {
	rows, cols := neovim.CalculateGridSize(width, height)
	if neovim.NvimScreen != nil {
		neovim.NvimScreen.Resize(cols, rows)
	}
}

// RequestState triggers emission of current state to frontend
func (a *App) RequestState() {
	if neovim.NvimScreen != nil {
		neovim.NvimScreen.EmitCurrentState()
	}
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
