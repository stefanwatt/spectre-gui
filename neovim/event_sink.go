package neovim

import (
	"sync/atomic"

	"nvim-gui/core/events"
	"nvim-gui/core/ports"
)

var canonicalEventSink atomic.Value

func SetEventSink(sink ports.EventSink) {
	if sink == nil {
		return
	}
	canonicalEventSink.Store(sink)
}

func EnqueueCanonicalEvent(event events.Event) {
	sink := canonicalEventSink.Load()
	if sink == nil {
		return
	}
	sink.(ports.EventSink).Enqueue(event)
}
