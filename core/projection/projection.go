package projection

import (
	"nvim-gui/core/model"
)

type UIProjection struct {
	Events []EmittedEvent
}

type EmittedEvent struct {
	Name    string
	Payload any
}

type Projector interface {
	Project(state model.AppState) UIProjection
}
