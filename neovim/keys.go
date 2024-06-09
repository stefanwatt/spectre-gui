package neovim

import (
	"fmt"
	"log"

	"spectre-gui/utils"

	"github.com/neovim/go-client/nvim"
)

var specialKeys = map[string]string{
	"Backspace":  "<BS>",
	"Enter":      "<CR>",
	"Escape":     "<Esc>",
	"Tab":        "<Tab>",
	"Insert":     "<Insert>",
	"Delete":     "<Del>",
	"ArrowUp":    "<Up>",
	"ArrowDown":  "<Down>",
	"ArrowLeft":  "<Left>",
	"ArrowRight": "<Right>",
	"Home":       "<Home>",
	"End":        "<End>",
	"PageUp":     "<PageUp>",
	"PageDown":   "<PageDown>",
	"F1":         "<F1>",
	"F2":         "<F2>",
	"F3":         "<F3>",
	"F4":         "<F4>",
	"F5":         "<F5>",
	"F6":         "<F6>",
	"F7":         "<F7>",
	"F8":         "<F8>",
	"F9":         "<F9>",
	"F10":        "<F10>",
	"F11":        "<F11>",
	"F12":        "<F12>",
	"Space":      "<Space>",
}

var ignored_keys = []string{
	"Super",
	"Alt",
	"Shift",
	"Control",
}

func SendKey(key string, alt bool, shift bool, ctrl bool, servername string) error {
	_, err := utils.Find(ignored_keys, func(ignored_key string) bool {
		return key == ignored_key
	})
	if err == nil {
		return fmt.Errorf("Ignored key: %s", key)
	}
	v, err := nvim.Dial(servername)
	if err != nil {
		log.Println("Failed to connect to Neovim:", err)
		return err
	}
	defer v.Close()

	if termcode, ok := specialKeys[key]; ok {
		key = termcode
	}

	sequence := ""
	if ctrl || alt || shift {
		sequence += "<"
		if ctrl {
			sequence += "C-"
		}
		if alt {
			sequence += "A-"
		}
		if shift {
			sequence += "S-"
		}
		sequence += key + ">"
	} else {
		sequence = key
	}

	_, err = v.Input(sequence)
	if err != nil {
		log.Println("Error feeding keys to Neovim:", err)
		return err
	}

	utils.LogTime("Sent key to Neovim: " + sequence)
	return nil
}
