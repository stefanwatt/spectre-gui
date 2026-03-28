package ports

import "nvim-gui/core/events"

type UIEmitter interface {
	Emit(name string, payload any)
}

type EventSink interface {
	Enqueue(event events.Event)
}

type Transport interface {
	Start() error
	Stop() error
}
