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
	GetBufferLines(buf int, start, end int, strict bool) ([][]byte, error)
	Command(cmd string) error
	OpenSplitRight(winId *int, bufNr int) error
	CurrentWindow() (int, error)
	SetWindowCursor(winId, row, col int) error
	SetCurrentWindow(winId int) error
	SetBufferToWindow(winId int, bufNr int) error
	GetCurrentFilepath() (string, error)
	SetWindowOption(winId int, key string, value any) error
	CreateBufferKeymap(bufNr int, mode, lhs string, rhs func(channelID int) string) error
	CreateBufferAutocmd(winId, bufNr int, luaCallback string) error
	AttachBuffer(bufNr int, sendBuffer bool, opts map[string]any) (bool, error)
	DetachBuffer(bufNr int) (bool, error)
	RegisterHandler(event string, handler func(data ...any))
}
