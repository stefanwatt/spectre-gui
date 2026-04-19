package neovim

import (
	"nvim-gui/rendering"
	neovimtransport "nvim-gui/transport/neovim"

	"nvim-gui/core/events"

	"github.com/charmbracelet/log"
	"github.com/neovim/go-client/nvim"
)

func HandleRedraw(updates [][]interface{}) {
	mapped := neovimtransport.MapRedrawBatch(updates)
	if len(mapped) > 0 {
		hasFlush := false
		for _, ev := range mapped {
			if ev.Name == events.EventFlush {
				hasFlush = true
				break
			}
		}
		if hasFlush {
			counts := make(map[events.Name]int)
			for _, ev := range mapped {
				counts[ev.Name]++
			}
			log.Debug("redraw batch mapped", "raw_updates", len(updates), "mapped_events", len(mapped), "counts", counts)
		}
	}

	for _, ev := range mapped {
		handleCanonicalSideEffects(ev)
		EnqueueCanonicalEvent(ev)
	}
}

func handleCanonicalSideEffects(ev events.Event) {
	switch ev.Name {
	case events.EventDefaultColorsSet:
		payload, ok := ev.Payload.(events.DefaultColorsSet)
		if !ok {
			return
		}
		DefaultColorsSet(payload.FG, payload.BG, payload.SP)
	case events.EventHighlightDefine:
		payload, ok := ev.Payload.(events.HighlightAttrDefine)
		if !ok {
			return
		}
		args := make([]interface{}, 0, len(payload.Args))
		for _, arg := range payload.Args {
			iface, ok := arg.([]interface{})
			if !ok {
				continue
			}
			args = append(args, iface)
		}
		HlAttrDefine(args)
		css := rendering.BuildHighlightCSS()
		EmitEvent("highlight-css", css)

	case events.EventWinPos:
		payload, ok := ev.Payload.(events.WindowPosition)
		if !ok || payload.WindowID == 0 {
			return
		}
		go fetchAndEnqueueBufferInfo(payload.WindowID)
		// go fetchAndEnqueueWindowOptions(payload.WindowID)

	case events.EventWinFloatPos:
		payload, ok := ev.Payload.(events.FloatingWindowPosition)
		if !ok || payload.WindowID == 0 {
			return
		}
		go fetchAndEnqueueBufferInfo(payload.WindowID)
		// go fetchAndEnqueueWindowOptions(payload.WindowID)
	}
}

// fetchAndEnqueueWindowOptions fetches number/relativenumber for a window and enqueues
// an EventWindowOptions event to update WindowState.LineNumbers.
func fetchAndEnqueueWindowOptions(winId int) {
	win := nvim.Window(winId)
	var number, relNumber interface{}
	_ = NvimClient.WindowOption(win, "number", &number)
	_ = NvimClient.WindowOption(win, "relativenumber", &relNumber)
	log.Infof("[fetchAndEnqueueWindowOptions] winId=%d number=%v relativenumber=%v", winId, number, relNumber)
	EnqueueCanonicalEvent(events.Event{
		Name: events.EventWindowOptions,
		Payload: events.WindowOptions{
			WindowID:    winId,
			LineNumbers: true,
		},
		Source: "neovim-api",
	})
}

// fetchAndEnqueueBufferInfo calls GetWindowBuffer asynchronously and enqueues
// an EventWindowBufferInfo event with the result. This enriches window state
// with buffer metadata (filetype, filepath) that is not part of the Neovim
// redraw protocol.
func fetchAndEnqueueBufferInfo(windowID int) {
	buf, err := GetWindowBuffer(windowID)
	if err != nil {
		log.Debug("fetchAndEnqueueBufferInfo: failed", "windowID", windowID, "err", err)
		return
	}
	EnqueueCanonicalEvent(events.Event{
		Name: events.EventWindowBufferInfo,
		Payload: events.WindowBufferInfo{
			WindowID: windowID,
			Filetype: buf.Filetype,
			Filepath: buf.Filepath,
		},
		Source: "neovim-api",
	})
}
