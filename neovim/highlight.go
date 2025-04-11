package neovim

import (
	"fmt"
	"strings"
)

type Highlight struct {
	Foreground    int
	Background    int
	Special       int
	Reverse       bool
	Italic        bool
	Bold          bool
	Underline     bool
	Undercurl     bool
	Strikethrough bool
}

// tailwind classes for font stuff
var (
	BOLD           = "font-bold"
	ITALIC         = "italic"
	UNDERLINE      = "underline"
	UNDERCURL      = "underline decoration-wavy"
	STRIKETHROUGH  = "line-trough"
	CLASSES_IN_CSS = "/home/stefan/.config/nvim-gui/classes-in-css.txt"
	CSS_FILE       = "/home/stefan/.config/nvim-gui/nvim-hl.css"
)

func mapClassesString(classes []string) string {
	return strings.Join(classes, "-")
}

func (h *Highlight) toString() string {
	return fmt.Sprintf("fg: %d, bg: %d", h.Foreground, h.Background)
}

func (h *Highlight) fgHex() string {
	return fmt.Sprintf("#%06x", h.Foreground)
}
func (h *Highlight) bgHex() string {
	hex := fmt.Sprintf("#%06x", h.Background)
	if hex == "#000000" {
		return ""
	}
	return hex
}

func (h *Highlight) getClasses() []string {
	var classes []string

	fgHex := h.fgHex()
	if fgHex != "" { // Check if fgHex is not empty
		fgClass, exists := fgColorClasses[fgHex]
		if !exists {
			err := addForegroundColorClass(fgHex)
			assert(err == nil, fmt.Sprintf("failed to add fg color class for %s", fgHex))
			fgClass, exists = fgColorClasses[fgHex]
			assert(exists, fmt.Sprintf("failed to find fg color class for %s after adding", fgHex))
		}
		classes = append(classes, fgClass)
	}

	bgHex := h.bgHex()
	if bgHex != "" { // Check if bgHex is not empty
		bgClass, exists := bgColorClasses[bgHex]
		if !exists {
			err := addBackgroundColorClass(bgHex)
			assert(err == nil, fmt.Sprintf("failed to add bg color class for %s", bgHex))
			bgClass, exists = bgColorClasses[bgHex]
			assert(exists, fmt.Sprintf("failed to find bg color class for %s after adding", bgHex))
		}
		classes = append(classes, bgClass)
	}

	if h.Bold {
		classes = append(classes, BOLD)
	}

	if h.Italic {
		classes = append(classes, ITALIC)
	}

	if h.Underline {
		classes = append(classes, UNDERLINE)
	}

	if h.Undercurl {
		classes = append(classes, UNDERCURL)
	}

	if h.Strikethrough {
		classes = append(classes, STRIKETHROUGH)
	}

	return classes
}

func updateHighlightCSS(optionalData ...interface{}) {
	lines, err := readLinesFromFile(CLASSES_IN_CSS)
	classesInCss := make(map[string]bool)
	for _, line := range lines {
		classesInCss[strings.TrimSpace(line)] = true
	}
	if err != nil {
		panic(err.Error())
	}

	var cssBuilder strings.Builder
	var classesInCssBuilder strings.Builder
	for color, class := range fgColorClasses {
		if _, exists := classesInCss[class]; exists {
			continue
		}
		classesInCssBuilder.WriteString(class + "\n")
		cssBuilder.WriteString(fmt.Sprintf(`
				.%s {
					color: %s;
				}
			`, class, color))
	}
	for color, class := range bgColorClasses {
		if _, exists := classesInCss[class]; exists {
			continue
		}
		classesInCssBuilder.WriteString(class + "\n")
		cssBuilder.WriteString(fmt.Sprintf(`
				.%s {
					background-color: %s;
				}
			`, class, color))
	}
	err = appendStringToFile(CSS_FILE, cssBuilder.String())
	if err != nil {
		panic("could not write css file:\n" + err.Error())
	}
	err = appendStringToFile(CLASSES_IN_CSS, classesInCssBuilder.String())
	if err != nil {
		panic("could not write classes-in-css file:\n" + err.Error())
	}
}
