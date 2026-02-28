package neovim

import (
	"context"
	"fmt"
	"strings"

	Runtime "github.com/wailsapp/wails/v2/pkg/runtime"
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
	HasForeground bool
	HasBackground bool
}

// tailwind classes for font stuff
var (
	BOLD          = "font-bold"
	ITALIC        = "italic"
	UNDERLINE     = "underline"
	UNDERCURL     = "underline decoration-wavy"
	STRIKETHROUGH = "line-through"
)

func mapClassesString(classes []string) string {
	return strings.Join(classes, "-")
}

func (h *Highlight) toString() string {
	return fmt.Sprintf("fg: %d, bg: %d", h.Foreground, h.Background)
}

func (h *Highlight) fgHex() string {
	if !h.HasForeground {
		return ""
	}
	return fmt.Sprintf("#%06x", h.Foreground)
}

func (h *Highlight) bgHex() string {
	if !h.HasBackground {
		return ""
	}
	return fmt.Sprintf("#%06x", h.Background)
}

func (h *Highlight) getClasses() []string {
	var classes []string

	fgHex := h.fgHex()
	if fgHex != "" {
		fgColorClassesMu.Lock()
		fgClass, exists := fgColorClasses[fgHex]
		fgColorClassesMu.Unlock()
		if !exists {
			addForegroundColorClass(fgHex)
			fgColorClassesMu.Lock()
			fgClass = fgColorClasses[fgHex]
			fgColorClassesMu.Unlock()
		}
		classes = append(classes, fgClass)
	}

	bgHex := h.bgHex()
	if bgHex != "" {
		bgColorClassesMu.Lock()
		bgClass, exists := bgColorClasses[bgHex]
		bgColorClassesMu.Unlock()
		if !exists {
			addBackgroundColorClass(bgHex)
			bgColorClassesMu.Lock()
			bgClass = bgColorClasses[bgHex]
			bgColorClassesMu.Unlock()
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

func emitHighlightCSS(ctx context.Context) {
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

	Runtime.EventsEmit(ctx, "highlight-css", cssBuilder.String())
}
