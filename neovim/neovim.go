package neovim

import (
	"context"
	"fmt"
	"log"
	"slices"

	"spectre-gui/utils"

	"github.com/neovim/go-client/nvim"
	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
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

func SendKey(key string, ctrl bool, alt bool, shift bool, servername string) error {
	if key == "Super" {
		return fmt.Errorf("Invalid key: %s", key)
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

	return nil
}

type BufChangeEvent struct {
	Event             string
	Buffer            int64
	ChangedTick       int64
	FirstLine         int64
	LastLine          int64
	LastLineChanged   int64
	PreviousByteCount int64
	DeletedCodepoints *int64

	DeletedCodeunits *int64
}

type HighlightToken struct {
	Text          string `msgpack:"text" json:"text"`
	StartRow      uint64 `msgpack:"start_row" json:"end_row"`
	EndRow        uint64 `msgpack:"end_row" json:"end_row"`
	StartCol      uint64 `msgpack:"start_col" json:"start_col"`
	EndCol        uint64 `msgpack:"end_col" json:"end_col"`
	Foreground    string `msgpack:"foreground" json:"foreground"`
	Background    string `msgpack:"background" json:"background"`
	Reverse       bool   `msgpack:"reverse" json:"reverse"`
	Underline     bool   `msgpack:"underline" json:"underline"`
	Undercurl     bool   `msgpack:"undercurl" json:"undercurl"`
	Strikethrough bool   `msgpack:"strikethrough" json:"strikethrough"`
	Bold          bool   `msgpack:"bold" json:"bold"`
	Italic        bool   `msgpack:"italic" json:"italic"`
}

type BufLine struct {
	Sign   string           `msgpack:"sign" json:"sign"`
	Row    uint64           `msgpack:"row" json:"row"`
	Tokens []HighlightToken `msgpack:"tokens" json:"tokens"`
}

func OnBufChanged(ctx context.Context, v *nvim.Nvim, args []interface{}) {
	var hl_tokens []HighlightToken
	nvim_cmd := "return require('config.nvim-gui').get_tokens(0,0,100,1)"
	log.Println("getting buf lines")
	err := v.ExecLua(nvim_cmd, &hl_tokens)
	if err != nil {
		utils.Log(err.Error())
		return
	}
	slices.SortFunc(hl_tokens, func(a HighlightToken, b HighlightToken) int {
		return int(a.StartRow - b.StartRow)
	})
	buf_lines := map_buf_lines(hl_tokens)
	Runtime.EventsEmit(ctx, "buf-lines-changed", buf_lines)
}

func map_buf_lines(tokens []HighlightToken) []BufLine {
	var buf_lines []BufLine
	row := 0
	var tokens_of_line []HighlightToken
	for _, token := range tokens {
		if row != int(token.StartRow) {
			buf_lines = append(buf_lines, BufLine{
				Sign:   "",
				Row:    uint64(row),
				Tokens: tokens_of_line,
			})
			tokens_of_line = []HighlightToken{}
			row++
		}
		for row < int(token.StartRow) {
			buf_lines = append(buf_lines, BufLine{
				Sign:   "",
				Row:    uint64(row),
				Tokens: []HighlightToken{},
			})
			tokens_of_line = []HighlightToken{}
			row++
		}
		tokens_of_line = append(tokens_of_line, token)
	}
	return buf_lines
}

func parse_lua_number(value interface{}) int {
	switch value := value.(type) {
	case int64:
		return int(value)
	case uint64:
		return int(value)
	default:
		return 0
	}
}

type CursorMoveEvent struct {
	Row        uint64 `msgpack:"row" json:"row"`
	Col        uint64 `msgpack:"col" json:"col"`
	Key        string `msgpack:"key" json:"key"`
	TopLine    uint64 `msgpack:"top_line" json:"top_line"`
	BottomLine uint64 `msgpack:"bottom_line" json:"bottom_line"`
}

func StartListening(servername string, ctx context.Context) {
	v, err := nvim.Dial(servername)
	if err != nil {
		log.Println(err)
		return
	}
	defer v.Close()
	OnBufChanged(ctx, v, nil)
	var result string
	nvim_cmd := fmt.Sprintf("return require('config.nvim-gui').attach_buffer(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	nvim_cmd = fmt.Sprintf("return require('config.nvim-gui').listen_for_cursor_move(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	nvim_cmd = fmt.Sprintf("return require('config.nvim-gui').listen_for_mode_change(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}

	v.RegisterHandler("nvim-gui-current-buf-changed", func(v *nvim.Nvim, args []interface{}) {
		OnBufChanged(ctx, v, args)
	})

	v.RegisterHandler("nvim-gui-mode-changed", func(v *nvim.Nvim, args []string) {
		Runtime.EventsEmit(ctx, "mode-changed", args[0])
	})

	v.RegisterHandler("nvim-gui-cursor-moved", func(v *nvim.Nvim, cursor_move_event CursorMoveEvent) {
		log.Println("cursor moved ", cursor_move_event)
		if cursor_move_event.Key == "" {
			cursor_move_event.Key = " "
		}
		Runtime.EventsEmit(ctx, "cursor-changed", cursor_move_event)
	})

	if err := v.Serve(); err != nil {
		log.Fatal(err)
	}
	log.Println("listening terminating")
}

func OpenFileAt(path string, row int, col int, servername string) error {
	utils.Log("[NEOVIM] Opening file at", path, row, col)
	v, err := nvim.Dial(servername)
	if err != nil {
		log.Println(err)
		return err
	}
	defer v.Close()
	err = v.Command(fmt.Sprintf("e %s", path))
	if err != nil {
		log.Println("Error opening file:", err)
		return err
	}

	err = v.Command(fmt.Sprintf("call cursor(%d, %d)", row, col))
	if err != nil {
		log.Println("Error setting cursor:", err)
		return err
	}
	return nil
}
