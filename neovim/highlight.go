package neovim

import "fmt"

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

func (h *Highlight)toString()string{
		return fmt.Sprintf("fg: %d, bg: %d",h.Foreground, h.Background)
}

func (h *Highlight) fgHex()string{
	return fmt.Sprintf("#%06x", h.Foreground)
}
func (h *Highlight) bgHex()string{
	return fmt.Sprintf("#%06x", h.Background)
}
