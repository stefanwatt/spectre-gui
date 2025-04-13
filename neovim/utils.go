package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"os"
	"path/filepath"
	"strings"
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
