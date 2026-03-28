package main

import (
	"fmt"
	"nvim-gui/features"
	ext "nvim-gui/features/picker/external-tools"
	"nvim-gui/neovim"

	"github.com/charmbracelet/log"
)

func (a *App) Paste(text string) {
	err := neovim.Paste(text)
	if err != nil {
		log.Error(err.Error())
	}
}

func (a *App) SendKey(key string, ctrl bool, alt bool, shift bool, keymapMode string) {
	log.Debug(fmt.Sprintf("SendKey key=%s ctrl=%t alt=%t shift=%t mode=%s", key, ctrl, alt, shift, keymapMode))
	switch keymapMode {
	// case "cmdline":
	default:
		err := neovim.SendKey(key, ctrl, alt, shift)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func (a *App) SubstituteJump() {
	features.HandleSubstituteJump(features.CmdlineBridge{
		GetCmdline: neovim.GetCmdline,
		GetCmdpos:  neovim.GetCmdpos,
		SetCmdline: neovim.SetCmdline,
	})
}

func (a *App) CreateQuickfixList(entries []*neovim.QuickfixEntry) {
	neovim.SetQuickfixList(entries)
}

func (a *App) GetReplacementText(matchedLine string, searchTerm string, replacementText string, useRegex bool) string {
	replacementText, err := ext.GetReplacementText(matchedLine, searchTerm, replacementText, useRegex)
	if err != nil {
		return ""
	}
	return replacementText
}
