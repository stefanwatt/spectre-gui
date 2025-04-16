package neovim

import (
	"fmt"
	"log"
	"strings"

	"nvim-gui/utils"

	"github.com/neovim/go-client/nvim"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Keymaps struct {
	GuiFindFiles               string `msgpack:"gui_find_files"`
	GuiFindReferences          string `msgpack:"gui_find_references"`
	GuiLiveGrep                string `msgpack:"gui_live_grep"`
	GuiFindBufferSymbols       string `msgpack:"gui_find_buffer_symbols"`
	FzfLuaFindFiles            string `msgpack:"fzf_lua_find_files"`
	FzfLuaFindReferences       string `msgpack:"fzf_lua_find_references"`
	FzfLuaLiveGrep             string `msgpack:"fzf_lua_live_grep"`
	FzfLuaFindBufferSymbols    string `msgpack:"fzf_lua_find_buffer_symbols"`
	FzfLuaFindWorkspaceSymbols string `msgpack:"fzf_lua_find_workspace_symbols"`
	FzfLuaFindHelp             string `msgpack:"fzf_lua_find_help"`
	FzfLuaFindBuffer           string `msgpack:"fzf_lua_find_buffer"`
	FzfLuaFindProject          string `msgpack:"fzf_lua_find_project"`
	FzfLuaFindTodo             string `msgpack:"fzf_lua_find_todo"`
}

var keymaps = Keymaps{
	GuiFindFiles:               "<leader>ff",
	GuiFindReferences:          "<leader>fr",
	GuiFindBufferSymbols:       "<leader>fs",
	GuiLiveGrep:                "<leader>fw",
	FzfLuaFindFiles:            "<leader>fF",
	FzfLuaFindReferences:       "<leader>fR",
	FzfLuaLiveGrep:             "<leader>fW",
	FzfLuaFindBufferSymbols:    "<leader>fS",
	FzfLuaFindWorkspaceSymbols: "<leader><leader>fS",
	FzfLuaFindHelp:             "<leader>fh",
	FzfLuaFindBuffer:           "<leader>fb",
	FzfLuaFindProject:          "<leader>fp",
	FzfLuaFindTodo:             "<leader>ft",
}
var keymapOpts = map[string]bool{
	"noremap": false,
}

func RegisterKeymap(event string, lhs string, cb func(v *nvim.Nvim, data interface{})) {
	NvimInstance.SetKeyMap("n", lhs, fmt.Sprintf("<cmd>lua vim.fn.rpcrequest(%d,'%s',{})<cr>", NvimInstance.ChannelID(), event), keymapOpts)
	NvimInstance.RegisterHandler(event, cb)
}

func SetupKeymaps() {
	RegisterKeymap("live-grep", keymaps.GuiLiveGrep, func(_ *nvim.Nvim, data interface{}) {
		Runtime.EventsEmit(NvimScreen.ctx, "show_live_grep")
	})

	RegisterKeymap("find-files", keymaps.GuiFindFiles, func(_ *nvim.Nvim, data interface{}) {
		Runtime.EventsEmit(NvimScreen.ctx, "show-find-files")
	})

	RegisterKeymap("find-references", keymaps.GuiFindReferences, func(_ *nvim.Nvim, data interface{}) {
		Runtime.EventsEmit(NvimScreen.ctx, "show-find-references")
	})

	RegisterKeymap("find-buffer-symbols", keymaps.GuiFindBufferSymbols, func(_ *nvim.Nvim, data interface{}) {
		Runtime.EventsEmit(NvimScreen.ctx, "show-find-buffer-symbols")
	})
}

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

func SendKey(key string, ctrl bool, alt bool, shift bool) error {
	_, err := utils.Find(ignored_keys, func(ignored_key string) bool {
		return key == ignored_key
	})
	if err == nil {
		return fmt.Errorf("Ignored key: %s", key)
	}

	if termcode, ok := specialKeys[key]; ok {
		key = termcode
		if ctrl || alt || shift {
			key = strings.TrimLeft(key, "<")
			key = strings.TrimRight(key, ">")
		}
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
	utils.Log("SendKey sequence:", sequence)

	_, err = NvimInstance.Input(sequence)
	if err != nil {
		log.Println("Error feeding keys to Neovim:", err)
		return err
	}

	return nil
}
