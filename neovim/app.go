package neovim

import "github.com/wailsapp/wails/v3/pkg/application"

// App holds the reference to the Wails application instance
// This allows the neovim package to emit events without being a service itself
var App *application.App

// SetApp sets the application reference
func SetApp(app *application.App) {
	App = app
}
