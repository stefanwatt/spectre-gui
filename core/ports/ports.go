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

type NvimClient interface {
	CreateBuffer(listed, scratch bool) (int, error)
	SetBufferLines(buf int, start, end int, strict bool, lines [][]byte) error
	Command(cmd string) error
	OpenSplitRight(winId *int, bufNr int) error
	CurrentWindow() (int, error)
	SetBufferToWindow(winId int, bufNr int) error
	GetCurrentFilepath() (string, error)
	SetWindowOption(winId int, key string, value any) error
}
