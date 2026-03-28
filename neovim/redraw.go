package neovim

import (
	"nvim-gui/rendering"
	neovimtransport "nvim-gui/transport/neovim"

	"nvim-gui/core/events"

	"github.com/charmbracelet/log"
)

func HandleRedraw(updates [][]interface{}) {
	mapped := neovimtransport.MapRedrawBatch(updates)
	if len(mapped) > 0 {
		counts := make(map[events.Name]int)
		for _, ev := range mapped {
			counts[ev.Name]++
		}
		log.Debug("redraw batch mapped", "raw_updates", len(updates), "mapped_events", len(mapped), "counts", counts)
	}

	for _, ev := range mapped {
		handleCanonicalSideEffects(ev)
		applyMappedEventToScreen(ev)
		if ev.Name == events.EventFlush && NvimScreen != nil {
			log.Debug("redraw flush -> screen render")
			NvimScreen.Render()
		}
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
	}
}
