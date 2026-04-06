package neovim

import (
	"github.com/charmbracelet/log"
)

type FileExplorerConfirmPrompt struct {
	Message string   `json:"message"`
	Choices []string `json:"choices"`
}

func (s *Screen) HandleFileExplorerConfirmChoice(choice int) {
	if NvimClient == nil {
		return
	}
	// vim.fn.confirm with '&Yes\n&No' responds to 'y' (Yes) or 'n' (No).
	key := "n"
	if choice == 1 {
		key = "y"
	}
	if _, err := NvimClient.Input(key); err != nil {
		log.Debug("[minifiles] failed sending confirm input", "key", key, "err", err)
		return
	}
	EmitEvent("file-explorer-confirm-prompt-hide", struct{}{})
}
