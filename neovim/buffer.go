package neovim

import (
	"errors"
	"path/filepath"

	"github.com/neovim/go-client/nvim"
)

type Buffer struct {
	Filepath string
	Filetype string
}

func getWindowBuffer(winId int) (*Buffer, error) {
	windows, error := NvimInstance.Windows()
	if error != nil {
		return nil, error
	}

	var foundWindow *nvim.Window
	for _, window := range windows {
		currentWinId, _ := extractWindowId(window)
		if currentWinId == winId {
			foundWindow = &window
		}
	}
	if foundWindow == nil {
		return nil, errors.New("couldnt find window")
	}
	buffer, error := NvimInstance.WindowBuffer(*foundWindow)

	if error != nil {
		return nil, error
	}

	// var filetype string
	// error = NvimInstance.BufferOption(buffer, "filetype", &filetype)
	//
	// if error != nil {
	// 	return nil, error
	// }
	// if filetype == "" {
	// 	var treesitterContext bool
	// 	NvimInstance.WindowVar(*foundWindow, "treesitter_context", &treesitterContext)
	// 	filetype = "treesitter_context"
	// }
	//
	buffername, err := NvimInstance.BufferName(buffer)
	assert(err == nil, "error getting bufname")
	result := Buffer{
		Filetype: "",
			Filepath: filepath.Base(buffername),
		}
	return &result, nil
}
