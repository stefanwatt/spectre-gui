package neovim

import (
	"bufio"
	"fmt"
	"nvim-gui/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neovim/go-client/nvim"
)

func GetFileIcon(filename string) (string, string) {
	extension := strings.TrimLeft(filepath.Ext(filename), ".")
	var iconRes = struct {
		Icon string `msgpack:"icon"`
		Hl   string `msgpack:"hl"`
	}{}

	err := NvimInstance.ExecLua(
		fmt.Sprintf(
			"local icon, hl = require('nvim-web-devicons').get_icon('%s', '%s', {default=true}); return {icon=icon, hl=hl}",
			filename,
			extension,
		),
		&iconRes,
	)
	if err != nil {
		panic("could not get icon\n" + err.Error())
	}
	hl, err := NvimInstance.HLByName(iconRes.Hl, true)
	if err != nil {
		panic("could not get icon highlight\n" + err.Error())
	}
	highlight := Highlight{Foreground: hl.Foreground}
	iconColor := highlight.fgHex()
	return iconRes.Icon, iconColor
}

func GetCwd() string {
	var cwd string
	err := NvimInstance.ExecLua("return vim.fn.getcwd()", &cwd)
	if err != nil {
		panic("could not get neovim cwd")
	}
	return cwd
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
func OpenFileAt(path string, row int, col int) error {
	utils.Log("[NEOVIM] Opening file at", path, row, col)

	assert(FileExists(path), "tried to open file that doesnt exist: "+path)
	err := NvimInstance.Command(fmt.Sprintf("e %s", path))
	if err != nil {
		utils.Log("Error opening file:", err)
		return err
	}

	err = NvimInstance.Command(fmt.Sprintf("call cursor(%d, %d)", row, col))
	if err != nil {
		utils.Log("Error setting cursor:", err)
		return err
	}
	return nil
}

func ParseLuaNumber(value interface{}) int {
	switch value := value.(type) {
	case int64:
		return int(value)
	case uint64:
		return int(value)
	default:
		return 0
	}
}

var previewWin *nvim.Window

func ClosePreview(winId int) {
	if previewWin != nil {
		previewWinId, err := strconv.Atoi(strings.Split(previewWin.String(), ":")[1])
		assert(err == nil, "ClosePreview error getting previewWinId")
		if previewWinId == winId {
			err = NvimInstance.CloseWindow(*previewWin, true)
			assert(err == nil, "ClosePreview error closing preview window")
			previewWin = nil
			return
		}
	}
	windows, err := NvimInstance.Windows()
	assert(err == nil, "ClosePreview error getting windows")
	for _, win := range windows {
		previewWinId, err := strconv.Atoi(strings.Split(win.String(), ":")[1])
		assert(err == nil, "ClosePreview error getting previewWinId")
		if previewWinId == winId {
			err = NvimInstance.CloseWindow(win, true)
			assert(err == nil, "ClosePreview error closing preview window")
			previewWin = nil
			break
		}
	}
}

func ShowPreview(absolutePath string, startRow int, endRow int) {
	assert(FileExists(absolutePath), "tried to get highlighted content for file that doesnt exist: "+absolutePath)
	lines, err := ReadFileLines(absolutePath, startRow, endRow)
	assert(err == nil, "ShowPreview error reading lines")
	buf, err := NvimInstance.CreateBuffer(false, true)
	assert(err == nil, "ShowPreview error creating buffer")
	utils.Log(fmt.Sprintf("ShowPreview buffer created: %s", buf.String()))
	err = NvimInstance.SetBufferLines(buf, startRow-1, endRow-1, false, lines)
	assert(err == nil, "ShowPreview error setting bufferlines")
	bufId, err := strconv.Atoi(strings.Split(buf.String(), ":")[1])
	assert(err == nil, "ShowPreview error getting buf id")
	var filetype string
	filename := filepath.Base(absolutePath)
	cmd := fmt.Sprintf("return vim.filetype.match({buf=%d, filename='%s'})", bufId, filename)
	err = NvimInstance.ExecLua(cmd, &filetype)
	if err != nil {
		filetype = filepath.Ext(absolutePath)
	}
	err = NvimInstance.SetBufferOption(buf, "filetype", filetype)
	assert(err == nil, "ShowPreview error setting filetype")
	if previewWin == nil {
		win, err := NvimInstance.OpenWindow(buf, false, &nvim.WindowConfig{
			Relative: "editor",
			Row:      3,
			Col:      50,
			Width:    100,
			Height:   30,
			ZIndex:   69420,
		})
		assert(err == nil, "ShowPreview error opening window")
		previewWin = &win
	} else {
		err = NvimInstance.SetBufferToWindow(*previewWin, buf)
		assert(err == nil, "ShowPreview error setting buffer on window")
	}
}

func ReadFileLines(filename string, startRow, endRow int) ([][]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()
	if startRow < 1 {
		return nil, fmt.Errorf("startRow must be greater than or equal to 1")
	}
	if endRow < startRow {
		return nil, fmt.Errorf("endRow must be greater than or equal to startRow")
	}

	scanner := bufio.NewScanner(file)
	currentLine := 1
	for currentLine < startRow && scanner.Scan() {
		currentLine++
	}
	if currentLine < startRow {
		return nil, fmt.Errorf("file has fewer than %d lines", startRow)
	}
	var lines [][]byte
	for currentLine <= endRow && scanner.Scan() {
		lineBytes := make([]byte, len(scanner.Bytes()))
		copy(lineBytes, scanner.Bytes())
		lines = append(lines, lineBytes)
		currentLine++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	utils.Log(fmt.Sprintf("ReadFileLines lines:"))
	for _, line := range lines {
		utils.Log(string(line))
	}
	return lines, nil
}
