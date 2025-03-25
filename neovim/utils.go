package neovim

import (
	"fmt"
	"log"
	"nvim-gui/utils"
)

func OpenFileAt(path string, row int, col int, servername string) error {
	utils.Log("[NEOVIM] Opening file at", path, row, col)
	err := NvimInstance.Command(fmt.Sprintf("e %s", path))
	if err != nil {
		log.Println("Error opening file:", err)
		return err
	}

	err = NvimInstance.Command(fmt.Sprintf("call cursor(%d, %d)", row, col))
	if err != nil {
		log.Println("Error setting cursor:", err)
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
