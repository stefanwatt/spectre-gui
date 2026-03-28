package projection

import "nvim-gui/core/model"

type NoopProjector struct{}

func (NoopProjector) Project(state model.AppState) UIProjection {
	return UIProjection{Events: []EmittedEvent{{Name: "state-updated", Payload: struct{}{}}}}
}
