package runtime

import (
	"context"
	"sync"

	"nvim-gui/core/events"
	"nvim-gui/core/model"
	"nvim-gui/core/ports"
	"nvim-gui/core/projection"
	"nvim-gui/core/reducer"
)

type Runtime struct {
	state *model.AppState

	emitter   ports.UIEmitter
	projector projection.Projector

	events chan events.Event
	done   chan struct{}

	mu sync.RWMutex
}

func New(emitter ports.UIEmitter) *Runtime {
	return &Runtime{
		state:     model.NewAppState(),
		emitter:   emitter,
		projector: projection.NoopProjector{},
		events:    make(chan events.Event, 2048),
		done:      make(chan struct{}),
	}
}

func (r *Runtime) Start(ctx context.Context) {
	go func() {
		defer close(r.done)
		for {
			select {
			case ev := <-r.events:
				r.handle(ev)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *Runtime) SetEmitter(emitter ports.UIEmitter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emitter = emitter
}

func (r *Runtime) SetProjector(projector projection.Projector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.projector = projector
}

func (r *Runtime) Wait() {
	<-r.done
}

func (r *Runtime) Enqueue(event events.Event) {
	r.events <- event
}

func (r *Runtime) Snapshot() model.AppState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return *r.state
}

func (r *Runtime) handle(event events.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	reducer.Apply(r.state, event)

	if event.Name == events.EventFlush && r.emitter != nil {
		projector := r.projector
		if projector == nil {
			projector = projection.NoopProjector{}
		}
		ui := projector.Project(r.state)
		for _, emitted := range ui.Events {
			r.emitter.Emit(emitted.Name, emitted.Payload)
		}
	}
}
