package main

import (
	"log"
	"nvim-gui/neovim"
	"nvim-gui/utils"
)

var (
	INFO_LEVEL    = "info"
	SUCCESS_LEVEL = "success"
	WARNING_LEVEL = "warning"
	ERROR_LEVEL   = "error"
	DELETE        = "file-deleted"
	REPLACE       = "file-replaced"
	REPLACE_ALL   = "replaced-all"
	UNDO          = "undo"
	TOAST         = "toast"
	write_event   = REPLACE
	page_size     = 20
)

func (a *App) GetAppState() AppState {
	return a.State
}

func (a *App) OpenMatch(path string, row int, col int) {
	utils.Log("a.Servername")
	err := neovim.OpenFileAt(path, row, col, a.Servername)
	if err != nil {
		utils.Log(err.Error())
	}
}

func (a *App) AddMatchesToQuickfixList() {
}

func (a *App) SendKey(key string, ctrl bool, alt bool, shift bool) {
	err := neovim.SendKey(key, ctrl, alt, shift, a.Servername)
	if err != nil {
		log.Println(err)
	}
}

