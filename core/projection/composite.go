package projection

import "nvim-gui/core/model"

type CompositeProjector struct {
	Projectors []Projector
}

func (c CompositeProjector) Project(state *model.AppState) UIProjection {
	out := UIProjection{Events: []EmittedEvent{}}
	for _, p := range c.Projectors {
		if p == nil {
			continue
		}
		part := p.Project(state)
		out.Events = append(out.Events, part.Events...)
	}
	return out
}
