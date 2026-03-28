package neovim

import (
	"errors"

	"nvim-gui/utils"
)

func GetCmdline() (string, error) {
	if NvimClient == nil {
		return "", errors.New("nvim client is not initialized")
	}
	var line string
	if err := NvimClient.ExecLua("return vim.fn.getcmdline()", &line); err != nil {
		return "", err
	}
	return line, nil
}

func GetCmdpos() (int, error) {
	if NvimClient == nil {
		return 0, errors.New("nvim client is not initialized")
	}
	var pos interface{}
	if err := NvimClient.ExecLua("return vim.fn.getcmdpos()", &pos); err != nil {
		return 0, err
	}
	return utils.ReflectToInt(pos), nil
}

func SetCmdline(content string, pos int) error {
	if NvimClient == nil {
		return errors.New("nvim client is not initialized")
	}
	if pos < 1 {
		pos = 1
	}
	var ok bool
	return NvimClient.ExecLua(`
local line, p = ...
if type(line) ~= "string" then
  return false
end
if type(p) ~= "number" then
  return false
end
vim.fn.setcmdline(line, p)
return true
`, &ok, content, pos)
}
