package neovim

import (
	"fmt"

	"github.com/charmbracelet/log"
)

type FileExplorerConfirmPrompt struct {
	WinId   int      `json:"winId"`
	Message string   `json:"message"`
	Choices []string `json:"choices"`
}

func (s *Screen) HandleFileExplorerConfirmChoice(winId, choice int) {
	if NvimClient == nil || winId < 1 {
		return
	}
	key := "<Esc>"
	if choice == 1 {
		key = "<CR>"
	}
	lua := `
local win_id, feed = ...
if not vim.api.nvim_win_is_valid(win_id) then
  return false
end
vim.api.nvim_set_current_win(win_id)
vim.api.nvim_feedkeys(vim.api.nvim_replace_termcodes(feed, true, false, true), "n", false)
return true
`
	var ok bool
	if err := NvimClient.ExecLua(lua, &ok, winId, key); err != nil {
		log.Debug(fmt.Sprintf("[minifiles] failed sending confirm response to win=%d: %v", winId, err))
		return
	}
	if ok {
		EmitEvent("file-explorer-confirm-prompt-hide", struct{}{})
	}
}
