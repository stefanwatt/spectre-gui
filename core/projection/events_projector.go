package projection

import (
	"nvim-gui/core/model"
	"strings"
)

type EventsProjector struct{}

func (EventsProjector) Project(state *model.AppState) UIProjection {
	s := state.Editor.Screen
	result := UIProjection{Events: []EmittedEvent{{Name: "state-updated", Payload: struct{}{}}}}

	if s.Cmdline.Visible {
		result.Events = append(result.Events, EmittedEvent{
			Name: "cmdline_show",
			Payload: map[string]any{
				"content": buildCmdlineContent(s.Cmdline.Chunks),
				"pos":     s.Cmdline.Pos,
				"firstc":  s.Cmdline.Firstc,
				"prompt":  s.Cmdline.Prompt,
				"indent":  s.Cmdline.Indent,
			},
		})
	} else {
		result.Events = append(result.Events, EmittedEvent{Name: "cmdline_hide", Payload: struct{}{}})
	}

	result.Events = append(result.Events, EmittedEvent{
		Name: "cmdline_pos",
		Payload: map[string]any{
			"pos":   s.Cmdline.Pos,
			"level": s.Cmdline.Level,
		},
	})

	return result
}

func buildCmdlineContent(chunks []any) string {
	var b strings.Builder
	for _, raw := range chunks {
		chunk, ok := raw.([]interface{})
		if !ok || len(chunk) == 0 {
			continue
		}
		if len(chunk) < 2 {
			if s, ok := chunk[0].(string); ok {
				b.WriteString(strings.ReplaceAll(s, "\t", " "))
			}
			continue
		}
		if s, ok := chunk[1].(string); ok {
			b.WriteString(strings.ReplaceAll(s, "\t", " "))
		}
	}
	return b.String()
}
