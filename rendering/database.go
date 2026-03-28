package rendering

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
)

var (
	currentFgColorId = 1
	currentBgColorId = 1
	idClassesMu      sync.Mutex
	idClasses        = make(map[int][]string)
	fgColorClassesMu sync.Mutex
	fgColorClasses   = make(map[string]string)
	bgColorClassesMu sync.Mutex
	bgColorClasses   = make(map[string]string)
	effectiveHlIdsMu sync.Mutex
	effectiveHlIds   = make(map[string]int)
)

func UpdateEffectiveHlId(hlClassesStr string, hlId int) int {
	var effectiveHlId int
	var existsHlId bool
	effectiveHlIdsMu.Lock()
	if effectiveHlId, existsHlId = effectiveHlIds[hlClassesStr]; !existsHlId {
		effectiveHlIds[hlClassesStr] = hlId
		effectiveHlId = hlId
	}
	effectiveHlIdsMu.Unlock()
	return effectiveHlId
}

func BuildHighlightCSS() string {
	var cssBuilder strings.Builder

	fgColorClassesMu.Lock()
	for color, class := range fgColorClasses {
		cssBuilder.WriteString(fmt.Sprintf(".%s{color:%s}", class, color))
	}
	fgColorClassesMu.Unlock()

	bgColorClassesMu.Lock()
	for color, class := range bgColorClasses {
		cssBuilder.WriteString(fmt.Sprintf(".%s{background-color:%s}", class, color))
	}
	bgColorClassesMu.Unlock()

	return cssBuilder.String()
}

func AddBackgroundColorClass(color string) {
	if color == "" {
		return
	}

	bgColorClassesMu.Lock()
	defer bgColorClassesMu.Unlock()

	if _, exists := bgColorClasses[color]; exists {
		return
	}

	class := fmt.Sprintf("bg-%d", currentBgColorId)
	log.Debug(fmt.Sprintf("adding class %s for color %s", class, color))
	currentBgColorId++
	bgColorClasses[color] = class
}

func AddForegroundColorClass(color string) {
	if color == "" {
		return
	}

	fgColorClassesMu.Lock()
	defer fgColorClassesMu.Unlock()

	if _, exists := fgColorClasses[color]; exists {
		return
	}

	class := fmt.Sprintf("fg-%d", currentFgColorId)
	log.Debug(fmt.Sprintf("adding class %s for color %s", class, color))
	currentFgColorId++
	fgColorClasses[color] = class
}

func AddIdClasses(id int, classes []string) {
	idClassesMu.Lock()
	defer idClassesMu.Unlock()

	if _, exists := idClasses[id]; exists {
		return
	}

	idClasses[id] = classes
}
