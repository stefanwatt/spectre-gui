package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/neovim/go-client/nvim"
)

type Keymaps struct {
	GuiFileExplorer            string `msgpack:"gui_file_explorer"`
	GuiFindFiles               string `msgpack:"gui_find_files"`
	GuiFindReferences          string `msgpack:"gui_find_references"`
	GuiLiveGrep                string `msgpack:"gui_live_grep"`
	GuiFindBufferSymbols       string `msgpack:"gui_find_buffer_symbols"`
	GuiFindHelp                string `msgpack:"gui_find_help"`
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
	GuiFileExplorer:            "<leader>e",
	GuiFindFiles:               "<leader>ff",
	GuiFindReferences:          "<leader>fr",
	GuiFindBufferSymbols:       "<leader>fs",
	GuiLiveGrep:                "<leader>fw",
	GuiFindHelp:                "<leader>fh",
	FzfLuaFindFiles:            "<leader>fF",
	FzfLuaFindReferences:       "<leader>fR",
	FzfLuaLiveGrep:             "<leader>fW",
	FzfLuaFindBufferSymbols:    "<leader>fS",
	FzfLuaFindWorkspaceSymbols: "<leader><leader>fS",
	FzfLuaFindHelp:             "<leader>fH",
	FzfLuaFindBuffer:           "<leader>fb",
	FzfLuaFindProject:          "<leader>fp",
	FzfLuaFindTodo:             "<leader>ft",
}

var keymapOpts = map[string]bool{
	"noremap": false,
}

// NOTE: we register the keymaps in neovim like this because i dont want to implement my own chords
func RegisterKeymap(event, lhs string, cb func(v *nvim.Nvim, data interface{})) {
	NvimClient.SetKeyMap("n", lhs, fmt.Sprintf("<cmd>lua vim.fn.rpcrequest(%d,'%s',{})<cr>", NvimClient.ChannelID(), event), keymapOpts)
	NvimClient.RegisterHandler(event, cb)
}

func SetupKeymaps() {
	RegisterKeymap("file-explorer", keymaps.GuiFileExplorer, func(_ *nvim.Nvim, _ interface{}) {
		fe := GetFileExplorer()
		if fe == nil {
			log.Error("[file-explorer] not initialized")
			return
		}
		if err := fe.Open(nil); err != nil {
			log.Errorf("[file-explorer] open failed: %v", err)
			return
		}
		EmitEvent("show-file-explorer", struct{}{})
	})
	RegisterKeymap("live-grep", keymaps.GuiLiveGrep, func(_ *nvim.Nvim, data interface{}) {
		EmitEvent("show_live_grep", struct{}{})
	})

	RegisterKeymap("find-files", keymaps.GuiFindFiles, func(_ *nvim.Nvim, data interface{}) {
		EmitEvent("show-find-files", struct{}{})
	})

	RegisterKeymap("find-references", keymaps.GuiFindReferences, func(_ *nvim.Nvim, data interface{}) {
		EmitEvent("show-find-references", struct{}{})
	})

	RegisterKeymap("find-buffer-symbols", keymaps.GuiFindBufferSymbols, func(_ *nvim.Nvim, data interface{}) {
		EmitEvent("show-find-buffer-symbols", struct{}{})
	})

	RegisterKeymap("find-help", keymaps.GuiFindHelp, func(_ *nvim.Nvim, data interface{}) {
		EmitEvent("show-find-help", struct{}{})
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

var shiftedChars = map[string]bool{
	"!":  true, // Shift+1
	"@":  true, // Shift+2
	"#":  true, // Shift+3
	"$":  true, // Shift+4
	"%":  true, // Shift+5
	"^":  true, // Shift+6
	"&":  true, // Shift+7
	"*":  true, // Shift+8
	"(":  true, // Shift+9
	")":  true, // Shift+0
	"_":  true, // Shift+-
	"+":  true, // Shift+=
	"{":  true, // Shift+[
	"}":  true, // Shift+]
	"|":  true, // Shift+\
	":":  true, // Shift+;
	"\"": true, // Shift+'
	// "<":  true, // Shift+, //TODO: this fucks my indent keymap
	">": true, // Shift+.
	"?": true, // Shift+/
	"~": true, // Shift+`
}

func Paste(text string) error {
	_, err := NvimClient.Paste(text, true, -1)
	if err != nil {
		log.Error("Error pasting to Neovim:", err)
		return err
	}
	return nil
}

func SendKey(key string, ctrl, alt, shift bool) error {
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

	applyShift := shift
	if shift && len(key) == 1 && shiftedChars[key] {
		applyShift = false
	}

	sequence := ""
	if ctrl || alt || applyShift {
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
	log.Debug("SendKey sequence:", sequence)

	_, err = NvimClient.Input(sequence)
	if err != nil {
		log.Error("Error feeding keys to Neovim:", err)
		return err
	}

	return nil
}
