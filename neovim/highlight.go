package neovim

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
