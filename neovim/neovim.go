package neovim

import (
	"context"
	"fmt"
	"log"

	"spectre-gui/highlighting"
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

	actualKeys, err := v.ReplaceTermcodes(sequence, true, true, true)
	if err != nil {
		log.Println("Error replacing termcodes:", err)
		return err
	}

	err = v.FeedKeys(actualKeys, "t", true)
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
	DeletedCodeunits  *int64
}

func OnBufChanged(ctx context.Context, v *nvim.Nvim, args []interface{}) {
	lines_bytes, err := v.BufferLines(0, 0, -1, true)
	if err != nil {
		panic(err)
	}
	lines := utils.MapArray(lines_bytes, func(bytes []byte) string {
		s := string(bytes)
		html := highlighting.HighlightCode(s, "foo.go")
		return html
	})

	Runtime.EventsEmit(ctx, "buf-lines-changed", lines)
}

func parse_lua_number(value interface{}) int {
	switch value.(type) {
	case int64:
		row_int64, ok := value.(int64)
		if !ok {
			return 0
		}
		return int(row_int64)
	case uint64:
		row_uint64, ok := value.(uint64)
		if !ok {
			return 0
		}
		return int(row_uint64)
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
	var result string
	nvim_cmd := fmt.Sprintf("return require('config.utils').attach_buffer(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}
	log.Println("channel: ", v.ChannelID())
	nvim_cmd = fmt.Sprintf("return require('config.utils').listen_for_cursor_move(%d)", v.ChannelID())
	err = v.ExecLua(nvim_cmd, &result)
	if err != nil {
		utils.Log(err.Error())
	}

	v.RegisterHandler("nvim-gui-current-buf-changed", func(v *nvim.Nvim, args []interface{}) {
		OnBufChanged(ctx, v, args)
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
