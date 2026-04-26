# Code Context

## Files Retrieved
1. `core/ports/ports.go` (lines 18-34) - nvim interface boundary; tells which buffer/window ops features may call through adapter.
2. `neovim/nvim_adapter.go` (lines 10-97) - concrete adapter; shows exact wrappers that exist and which preview ops are missing here.
3. `neovim/utils.go` (lines 69-123, 149-193) - open-file path and file-preview path; scratch buffer, set lines, set filetype, open floating window.
4. `neovim/buffer.go` (lines 20-89) - buffer metadata read path; only getter for buffer name/filetype, no setter.
5. `neovim/neovim.go` (lines 85-117, 320-320) - redraw handler registration before `AttachUI`; global redraw stream setup.
6. `transport/neovim/redraw_mapper.go` (lines 155-281) - redraw event coverage for `win_pos`, `win_float_pos`, `win_hide`, `win_close`.
7. `neovim/redraw.go` (lines 38-117) - side effects from redraw: fetch buffer info on visible/floating window position events.
8. `features/picker/picker.go` (lines 86-103) - frontend entry points into preview/open-file flow.

## Key Code

### Interface / adapter coverage
`core/ports/ports.go` (18-34):
```go
type NvimClient interface {
	CreateBuffer(listed, scratch bool) (int, error)
	SetBufferLines(buf int, start, end int, strict bool, lines [][]byte) error
	GetBufferLines(buf int, start, end int, strict bool) ([][]byte, error)
	Command(cmd string) error
	OpenSplitRight(winId *int, bufNr int) error
	CurrentWindow() (int, error)
	SetWindowCursor(winId, row, col int) error
	SetCurrentWindow(winId int) error
	SetBufferToWindow(winId int, bufNr int) error
	GetCurrentFilepath() (string, error)
	SetWindowOption(winId int, key string, value any) error
	CreateBufferKeymap(bufNr int, mode, lhs string, rhs func(channelID int) string) error
	CreateBufferAutocmd(winId, bufNr int, luaCallback string) error
	AttachBuffer(bufNr int, sendBuffer bool, opts map[string]any) (bool, error)
	DetachBuffer(bufNr int) (bool, error)
	RegisterHandler(event string, handler func(data ...any))
}
```

`neovim/nvim_adapter.go` (14-97): adapter forwards only this subset. Preview-relevant wrappers exist for `CreateBuffer`, `SetBufferLines`, `GetBufferLines`, `Command`, `OpenSplitRight`, `SetBufferToWindow`, `AttachBuffer`, `DetachBuffer`.

### Preview path
`neovim/utils.go` (149-193):
```go
buf, err := NvimClient.CreateBuffer(false, true)
err = NvimClient.SetBufferLines(buf, 1, previewLines, false, lines)
err = NvimClient.ExecLua(cmd, &filetype)
err = NvimClient.SetBufferOption(buf, "filetype", filetype)
win, err := NvimClient.OpenWindow(buf, false, &nvim.WindowConfig{...})
// else: err = NvimClient.SetBufferToWindow(*previewWin, buf)
```
This is current preview implementation: Go reads file, Neovim gets scratch buffer + lines, then floating window. Buffer name/filepath not set here.

`neovim/utils.go` (69-84):
```go
err := NvimClient.Command(fmt.Sprintf("e %s", path))
err = NvimClient.Command(fmt.Sprintf("call cursor(%d, %d)", row, col))
```
Open-file path uses `:edit`, not a buffer API.

### Buffer metadata
`neovim/buffer.go` (36-89):
```go
buffername, err := NvimClient.BufferName(buffer)
result := Buffer{
	Filetype: filetype,
	Filepath: filepath.Base(buffername),
	BufNr:    bufNr,
}
```
Only read path exists. `Filepath` field is basename only.

### Redraw / hidden window path
`neovim/neovim.go` (85-117, 320-320): redraw handler registered before `AttachUI`, so redraw events include whole UI stream.

`transport/neovim/redraw_mapper.go` (197-281): maps `win_pos`, `win_float_pos`, `win_hide`, `win_close`.

`neovim/redraw.go` (63-77, 99-117):
- `WinPos` / `WinFloatPos` => async `fetchAndEnqueueBufferInfo()`
- `fetchAndEnqueueBufferInfo()` calls `GetWindowBuffer()` and emits `WindowBufferInfo`
- no fetch on `WinHide` / `WinClose`

### Picker entry points
`features/picker/picker.go` (90-103):
- `GetPreview()` -> `neovim.ShowPreview()`
- `OpenFile()` -> `neovim.OpenFileAt()`

## Architecture

- Preview flow is Go-owned file IO + Neovim scratch buffer.
- `CreateBuffer(false, true)` makes hidden/scratch buffer.
- `SetBufferLines` loads content; no `bufload`/`edit`/`SetBufferName` wrapper in adapter.
- `SetBufferOption` and `OpenWindow` are used directly on package-level `NvimClient`, outside `ports.NvimClient` / `NvimAdapter`.
- `AttachBuffer` in interface means `nvim_buf_attach` event subscription, not window attach.
- Window display uses `OpenWindow` first time, then `SetBufferToWindow` for reuse.
- Redraw is global via `AttachUI`; hidden/non-current window state comes from `redraw` events (`win_pos`, `win_float_pos`, `win_hide`, `win_close`). Buffer info is only re-fetched for visible/floating position events.

## Start Here
`neovim/utils.go` (lines 69-84, 98-193) - preview/open-file lifecycle lives here; fastest path to exact buffer ops and gaps.