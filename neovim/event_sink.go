package neovim

import (
	"nvim-gui/core/events"
	"nvim-gui/core/ports"
	"sync/atomic"

	fileexplorer "nvim-gui/features/file-explorer"
)

var (
	canonicalEventSink   atomic.Value
	fileExplorerRegistry atomic.Value
)

func SetFileExplorerRegistry(reg *fileexplorer.Registry) {
	fileExplorerRegistry.Store(reg)
}

func GetFileExplorerRegistry() *fileexplorer.Registry {
	v := fileExplorerRegistry.Load()
	if v == nil {
		return nil
	}
	return v.(*fileexplorer.Registry)
}

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
