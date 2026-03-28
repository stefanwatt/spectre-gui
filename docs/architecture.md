# Architecture Draft

This document defines the target architecture while the refactor is intentionally in a broken state.

## Layers

- `transport/*`: Inbound/outbound adapters for external systems (Neovim RPC, Wails runtime).
- `core/events`: Canonical event names and typed payloads.
- `core/model`: Canonical in-memory state.
- `core/reducer`: Pure state transitions from canonical events.
- `core/ports`: Interfaces for side effects and host integration.
- `features/*`: Feature-local state and feature-specific logic.
- `rendering/*`: Projection-only logic from canonical state to frontend payloads.
- `app/runtime`: Dependency graph, event loop, and lifecycle orchestration.

## Ownership Rules

- Core state lives in `core/model` only.
- Only reducers mutate core state.
- Rendering does not mutate state.
- Transport does not mutate state directly; it emits canonical events.
- Features own only feature-local state.
- Only runtime owns concrete adapters (Neovim, Wails).

## Forbidden Dependencies

- `core/*` must not import `github.com/wailsapp/wails/*`.
- `core/*` must not import `github.com/neovim/go-client/*`.
- `rendering/*` must not import transport packages.
- `features/*` must not import runtime packages.

## Event Flow

1. Transport receives raw events from Neovim.
2. Transport maps raw events to canonical `core/events` events.
3. Runtime dispatches canonical events through the reducer.
4. Runtime asks renderers/projectors for UI payloads.
5. Runtime emits payloads through `core/ports.UIEmitter`.

## Runtime Model

- Single owner goroutine for mutable app state.
- External callbacks enqueue events.
- Reducers are deterministic and side-effect free.
- Side effects are triggered outside reducers via ports.

## Migration Checklist

- Move event names/payloads to `core/events`.
- Move state structs from `neovim/screen.go` into `core/model`.
- Replace direct `App.Event.Emit` calls with `core/ports.UIEmitter`.
- Replace package globals (`NvimClient`, `NvimScreen`, `App`) with runtime wiring.
- Keep transport translation and state mutation separate.
