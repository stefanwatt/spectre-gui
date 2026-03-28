package neovim

import (
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type EventEmitter interface {
	Emit(eventName string, data interface{})
}

type wailsEventEmitter struct {
	app *application.App
}

func (w wailsEventEmitter) Emit(eventName string, data interface{}) {
	if w.app == nil {
		return
	}
	w.app.Event.Emit(eventName, data)
}

var eventEmitter atomic.Value

func SetEventEmitter(emitter EventEmitter) {
	if emitter == nil {
		return
	}
	eventEmitter.Store(emitter)
}

func EmitEvent(eventName string, data interface{}) {
	emitter := eventEmitter.Load()
	if emitter == nil {
		return
	}
	emitter.(EventEmitter).Emit(eventName, data)
}
