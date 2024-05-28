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

	err = v.FeedKeys(actualKeys, "n", true)
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
	event := BufChangeEvent{
		Event:           args[0].(string),
		Buffer:          args[1].(int64),
		ChangedTick:     args[2].(int64),
		FirstLine:       args[3].(int64),
		LastLine:        args[4].(int64),
		LastLineChanged: args[5].(int64),
	}

	if len(args) > 6 {
		if val, ok := args[6].(int64); ok {
			log.Println(args[6])
			event.DeletedCodepoints = &val
		}
	}
	if len(args) > 7 {
		if val, ok := args[7].(int64); ok {
			log.Println(args[7])
			event.DeletedCodeunits = &val
		}
	}
	// lines_bytes, err := v.BufferLines(0, int(event.FirstLine), int(event.LastLine), true)
	lines_bytes, err := v.BufferLines(0, 0, -1, true)
	if err != nil {
		panic(err)
	}
	lines := utils.MapArray(lines_bytes, func(bytes []byte) string {
		s := string(bytes)
		return highlighting.HighlightCode(s, "foo.go")
	})

	Runtime.EventsEmit(ctx, "buf-lines-changed", lines)
	log.Println("buf changed lines", lines)
	log.Printf("BufChanged: %+v\n", event)
}

func StartListening(servername string, ctx context.Context) {
	v, err := nvim.Dial(servername)
	var result string
	nvim_cmd := fmt.Sprintf("return require('config.utils').attach_buffer(%d)", v.ChannelID())
	log.Println("nvim attach cmd:", nvim_cmd)

	v.ExecLua(nvim_cmd, &result)
	if err != nil {
		log.Println(err)
	}
	log.Println("channel: ", v.ChannelID())
	defer v.Close()
	log.Println("start listening")

	v.RegisterHandler("spectre-current-buf-changed", func(v *nvim.Nvim, args []interface{}) {
		OnBufChanged(ctx, v, args)
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
