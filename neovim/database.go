package neovim

import (
	"fmt"
	"nvim-gui/utils"
	"sync"
)

var (
	currentFgColorId = 1
	currentBgColorId = 1
	fgColorClassesMu sync.Mutex
	bgColorClassesMu sync.Mutex
	idClassesMu      sync.Mutex
)

func addBackgroundColorClass(color string) {
	if color == "" {
		return
	}

	bgColorClassesMu.Lock()
	defer bgColorClassesMu.Unlock()

	if _, exists := bgColorClasses[color]; exists {
		return
	}

	class := fmt.Sprintf("bg-%d", currentBgColorId)
	utils.Log(fmt.Sprintf("adding class %s for color %s", class, color))
	currentBgColorId++
	bgColorClasses[color] = class
}

func addForegroundColorClass(color string) {
	if color == "" {
		return
	}

	fgColorClassesMu.Lock()
	defer fgColorClassesMu.Unlock()

	if _, exists := fgColorClasses[color]; exists {
		return
	}

	class := fmt.Sprintf("fg-%d", currentFgColorId)
	utils.Log(fmt.Sprintf("adding class %s for color %s", class, color))
	currentFgColorId++
	fgColorClasses[color] = class
}

func addIdClasses(id int, classes []string) {
	idClassesMu.Lock()
	defer idClassesMu.Unlock()

	if _, exists := idClasses[id]; exists {
		return
	}

	idClasses[id] = classes
}
