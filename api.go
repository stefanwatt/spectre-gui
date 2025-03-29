package main

import (
	"nvim-gui/neovim"
)

var (
	page_size     = 20
)

func (a *App) SendKey(key string, ctrl bool, alt bool, shift bool) {
	err := neovim.SendKey(key, ctrl, alt, shift)
	if err != nil {
		// log.Println(err)
	}
}

