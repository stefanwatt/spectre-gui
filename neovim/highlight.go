package neovim

import (
	"fmt"
	"nvim-gui/rendering"
	"nvim-gui/utils"
	"strings"
	"sync"
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
	highlights    = make(map[int]*Highlight)
	highlightsMu  sync.RWMutex // Mutex for Highlights map
)

func Init() {
	highlights[0] = &Highlight{
		Foreground: 0xffffff,
		Background: 0x000000,
		Special:    0xffffff,
	}
}

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
		fgClass := rendering.EnsureForegroundColorClass(fgHex)
		classes = append(classes, fgClass)
	}

	bgHex := h.bgHex()
	if bgHex != "" {
		bgClass := rendering.EnsureBackgroundColorClass(bgHex)
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

func DefaultColorsSet(fg, bg, sp int) {
	defaultFg := 0xffffff
	defaultBg := 0x000000
	defaultSp := defaultFg
	if fg >= 0 {
		defaultFg = fg
	}
	if bg >= 0 {
		defaultBg = bg
	}
	if sp >= 0 {
		defaultSp = sp
	}
	highlightsMu.Lock()
	highlights[0] = &Highlight{
		Foreground: defaultFg,
		Background: defaultBg,
		Special:    defaultSp,
	}
	highlightsMu.Unlock()
}

func HlAttrDefine(args []interface{}) {
	for _, attr := range args {
		attrData := attr.([]interface{})
		id := utils.ReflectToInt(attrData[0])
		rgbAttrs := attrData[1].(map[string]interface{})

		highlight := &Highlight{}

		if fg, ok := rgbAttrs["foreground"]; ok {
			highlight.Foreground = utils.ReflectToInt(fg)
			highlight.HasForeground = true
		}

		if bg, ok := rgbAttrs["background"]; ok {
			highlight.Background = utils.ReflectToInt(bg)
			highlight.HasBackground = true
		}

		if sp, ok := rgbAttrs["special"]; ok {
			highlight.Special = utils.ReflectToInt(sp)
		}

		if reverse, ok := rgbAttrs["reverse"]; ok {
			highlight.Reverse = reverse.(bool)
		}

		if italic, ok := rgbAttrs["italic"]; ok {
			highlight.Italic = italic.(bool)
		}

		if bold, ok := rgbAttrs["bold"]; ok {
			highlight.Bold = bold.(bool)
		}

		if underline, ok := rgbAttrs["underline"]; ok {
			highlight.Underline = underline.(bool)
		}

		if undercurl, ok := rgbAttrs["undercurl"]; ok {
			highlight.Undercurl = undercurl.(bool)
		}

		if strikethrough, ok := rgbAttrs["strikethrough"]; ok {
			highlight.Strikethrough = strikethrough.(bool)
		}

		highlightsMu.Lock()
		highlights[id] = highlight
		highlightsMu.Unlock()

		rendering.AddForegroundColorClass(highlight.fgHex())
		rendering.AddBackgroundColorClass(highlight.bgHex())

		hlClasses := highlight.getClasses()
		hlClassesStr := mapClassesString(hlClasses)
		effectiveHlId := rendering.UpdateEffectiveHlId(hlClassesStr, id)
		rendering.AddIdClasses(effectiveHlId, hlClasses)
	}
}
