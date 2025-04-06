package neovim

import (
	"fmt"
	"os"
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
	BOLD          = "font-bold"
	ITALIC        = "italic"
	UNDERLINE     = "underline"
	UNDERCURL     = "underline decoration-wavy"
	STRIKETHROUGH = "line-trough"
)

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
	if h.Foreground != 0 {
		fgClass, exists := colorClasses[h.fgHex()]
		// assert(exists, "tried to get classes for fg that wasnt added yet")
		if !exists {
			err := addColorClass(h.fgHex(), "fg")
			assert(err == nil, "failed to add color class")
			fgClass, exists = colorClasses[h.fgHex()]
			assert(exists, "failed to add color class")
		}
		classes = append(classes, fgClass)
	}
	if h.Background != 0 {
		bgClass, exists := colorClasses[h.bgHex()]
		if !exists {
			err := addColorClass(h.bgHex(), "bg")
			assert(err == nil, "failed to add color class")
			bgClass, exists = colorClasses[h.bgHex()]
			assert(exists, "failed to add color class")
		}
		// assert(exists, "tried to get classes for bg that wasnt added yet")
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
	var builder strings.Builder
	for color, class := range colorClasses {
		cssProperty := "color"
		if strings.HasPrefix(class, "bg") {
			cssProperty = "background-color"
		}
		builder.WriteString(fmt.Sprintf(`
				.%s {
					%s: %s;
				}
			`, class, cssProperty, color))
	}

	err := os.WriteFile("/home/stefan/.config/nvim-gui/nvim-hl.css", []byte(builder.String()), os.ModeAppend)
	assert(err == nil, "could not write css file")
}
