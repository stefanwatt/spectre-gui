package neovim

import (
	"github.com/charmbracelet/log"
	"github.com/neovim/go-client/nvim"
)

// NvimAdapter wraps the package-level NvimClient and implements ports.NvimClient.
// Used to inject nvim operations into features without import cycles.
type NvimAdapter struct{}

func (a *NvimAdapter) CreateBufferKeymap(bufNr int, mode, lhs string, _rhs func(channelID int) string) error {
	rhs := _rhs(NvimClient.ChannelID())
	log.Infof("[CreateBufferKeymap] rhs=%s", rhs)
	return NvimClient.SetBufferKeyMap(nvim.Buffer(bufNr), mode, lhs, rhs, map[string]bool{})
}
func (a *NvimAdapter) CreateBuffer(listed, scratch bool) (int, error) {
	buf, err := NvimClient.CreateBuffer(listed, scratch)
	return int(buf), err
}

func (a *NvimAdapter) SetBufferLines(bufNr int, start, end int, strict bool, lines [][]byte) error {
	return NvimClient.SetBufferLines(nvim.Buffer(bufNr), start, end, strict, lines)
}

func (a *NvimAdapter) Command(cmd string) error {
	return NvimClient.Command(cmd)
}

func (a *NvimAdapter) OpenSplitRight(winId *int, bufNr int) error {
	return NvimClient.ExecLua(`
			local bufNr = ...
			return vim.api.nvim_open_win(bufNr, true, { split = 'right', win = 0, })
		`, winId, bufNr)
}

func (a *NvimAdapter) CurrentWindow() (int, error) {
	win, err := NvimClient.CurrentWindow()
	return int(win), err
}
func (a *NvimAdapter) SetCurrentWindow(winId int) error {
	return NvimClient.SetCurrentWindow(nvim.Window(winId))
}

func (a *NvimAdapter) SetBufferToWindow(winId int, bufNr int) error {
	return NvimClient.SetBufferToWindow(nvim.Window(winId), nvim.Buffer(bufNr))
}

func (a *NvimAdapter) GetCurrentFilepath() (string, error) {
	var dir string
	err := NvimClient.ExecLua(`return vim.fn.expand("%:p")`, &dir)
	return dir, err
}

func (a *NvimAdapter) SetWindowOption(winId int, key string, value any) error {
	win := nvim.Window(winId)
	log.Infof("[FileExplorer] converted win: %v", win)
	return NvimClient.SetWindowOption(win, key, value)
}
