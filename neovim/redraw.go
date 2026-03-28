package neovim

import (
	"nvim-gui/rendering"
	neovimtransport "nvim-gui/transport/neovim"

	"nvim-gui/core/events"
)

func HandleRedraw(updates [][]interface{}) {
	for _, ev := range neovimtransport.MapRedrawBatch(updates) {
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
	}
}
