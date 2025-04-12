package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"os"
)

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

	assert(FileExists(path),"tried to open file that doesnt exist: "+path)
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
