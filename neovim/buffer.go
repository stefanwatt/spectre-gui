package neovim

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/neovim/go-client/nvim"
)

type Buffer struct {
	Filepath string
	Filetype string
	BufNr    int
}

func GetCurrentBuffer() (*Buffer, error) {
	win, err := NvimClient.CurrentWindow()
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
	return buf, nil
}

func GetWindowBuffer(winId int) (*Buffer, error) {
	log.Debug(fmt.Sprintf("[minifiles] GetWindowBuffer: fetching windows for winId=%d", winId))
	windows, error := NvimClient.Windows()
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
	log.Debug(fmt.Sprintf("[minifiles] GetWindowBuffer: found window, fetching buffer for winId=%d", winId))
	buffer, error := NvimClient.WindowBuffer(*foundWindow)

	if error != nil {
		return nil, error
	}

	var filetype string
	error = NvimClient.BufferOption(buffer, "filetype", &filetype)

	if error != nil {
		return nil, error
	}
	log.Debug(fmt.Sprintf("[minifiles] GetWindowBuffer: winId=%d filetype=%s", winId, filetype))
	if filetype == "" {
		var treesitterContext bool
		NvimClient.WindowVar(*foundWindow, "treesitter_context", &treesitterContext)
		filetype = "treesitter_context"
	}

	buffername, err := NvimClient.BufferName(buffer)
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
