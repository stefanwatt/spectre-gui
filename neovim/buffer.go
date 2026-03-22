package neovim

import (
	"errors"
	"fmt"
	"nvim-gui/utils"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neovim/go-client/nvim"
)

type Buffer struct {
	Filepath string
	Filetype string
	BufNr    int
}
func GetCurrentBuffer()(*Buffer,error){
	win, err := NvimInstance.CurrentWindow()
	if err != nil {
		return nil, err
	}
	winId, err := extractWindowId(win)
	if err != nil {
		return nil, err
	}
	buf, err := GetWindowBuffer(winId)
	if err != nil {
		return nil, err
	}
	return buf,nil
}

func GetWindowBuffer(winId int) (*Buffer, error) {
	utils.Log(fmt.Sprintf("[minifiles] GetWindowBuffer: fetching windows for winId=%d", winId))
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
	utils.Log(fmt.Sprintf("[minifiles] GetWindowBuffer: found window, fetching buffer for winId=%d", winId))
	buffer, error := NvimInstance.WindowBuffer(*foundWindow)

	if error != nil {
		return nil, error
	}

	var filetype string
	error = NvimInstance.BufferOption(buffer, "filetype", &filetype)

	if error != nil {
		return nil, error
	}
	utils.Log(fmt.Sprintf("[minifiles] GetWindowBuffer: winId=%d filetype=%s", winId, filetype))
	if filetype == "" {
		var treesitterContext bool
		NvimInstance.WindowVar(*foundWindow, "treesitter_context", &treesitterContext)
		filetype = "treesitter_context"
	}

	buffername, err := NvimInstance.BufferName(buffer)
	assert(err == nil, "error getting bufname")

	// Extract buffer number from buffer.String() which returns "buffer:N"
	bufNr := 0
	parts := strings.Split(buffer.String(), ":")
	if len(parts) == 2 {
		bufNr, _ = strconv.Atoi(parts[1])
	}

	result := Buffer{
		Filetype: filetype,
		Filepath: filepath.Base(buffername),
		BufNr:    bufNr,
	}
	return &result, nil
}
