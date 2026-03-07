package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var osWindowMgr *OSWindowManager

type OSWindowManager struct {
	app     *application.App
	windows map[int]*application.WebviewWindow // winId → Wails window (external windows only)
	mu      sync.Mutex
}

func InitOSWindowManager(app *application.App) {
	osWindowMgr = &OSWindowManager{
		app:     app,
		windows: make(map[int]*application.WebviewWindow),
	}
}

func (m *OSWindowManager) CreateWindow(winId, gridId int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.windows[winId]; exists {
		return
	}

	utils.Log(fmt.Sprintf("OSWindowManager: creating OS window for winId=%d gridId=%d", winId, gridId))

	window := m.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            fmt.Sprintf("nvim-gui [%d]", winId),
		Width:            800,
		Height:           600,
		BackgroundColour: application.NewRGB(39, 42, 56),
		URL:              fmt.Sprintf("/?winId=%d&gridId=%d", winId, gridId),
	})
	m.windows[winId] = window
	window.Show()
}

// SetCurrentWindow sets the given Neovim window as the active window.
// Called from the frontend when a webview receives focus.
func SetCurrentWindow(winId int) {
	if NvimInstance == nil {
		return
	}

	utils.Log(fmt.Sprintf("SetCurrentWindow: focusing Neovim window %d", winId))

	
	nvimWin, err := getWindow(winId)
	if err != nil {
		utils.Log(fmt.Sprintf("SetCurrentWindow: failed to get window %d: %v", winId, err))
		return
	}

	err = NvimInstance.SetCurrentWindow(*nvimWin)
	if err != nil {
		utils.Log(fmt.Sprintf("SetCurrentWindow: failed to set current window %d: %v", winId, err))
	}
}

func (m *OSWindowManager) CloseWindow(winId int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if window, exists := m.windows[winId]; exists {
		utils.Log(fmt.Sprintf("OSWindowManager: closing OS window for winId=%d", winId))
		window.Close()
		delete(m.windows, winId)
	}
}
