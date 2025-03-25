package main

import (
	"log"
	"nvim-gui/neovim"
)

var (
	page_size     = 20
)

func (a *App) SendKey(key string, ctrl bool, alt bool, shift bool) {
	err := neovim.SendKey(key, ctrl, alt, shift, a.Servername)
	if err != nil {
		log.Println(err)
	}
}

