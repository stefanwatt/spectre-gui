package main

import (
	"fmt"
	"nvim-gui/neovim"
	ext "nvim-gui/picker/external-tools"
	"nvim-gui/utils"
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
	neovim.HandleSubstituteJump()
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
