package picker

import (
	"bytes"
	undo "nvim-gui/picker/undo"
	"os/exec"
	"strings"
)

var (
	undo_stack    = undo.UndoStack{}
	INFO_LEVEL    = "info"
	SUCCESS_LEVEL = "success"
	WARNING_LEVEL = "warning"
	ERROR_LEVEL   = "error"
	DELETE        = "file-deleted"
	REPLACE       = "file-replaced"
	REPLACE_ALL   = "replaced-all"
	UNDO          = "undo"
	TOAST         = "toast"
	write_event   = REPLACE
	page_size     = 20
)

type PickerResult struct {
	Filename     string `json:"filename"`
	RelativePath string `json:"relativePath"`
	AbsolutePath string `json:"absolutePath"`
	Icon         string `json:"icon"`
	IconColor    string `json:"iconColor"`
	Text         string `json:"text"`
	Row          int    `json:"row"`
	Col          int    `json:"col"`
}

func filterWithFzf(lines []string, query string) ([]string, error) {
	input := strings.Join(lines, "\n")
	fzf := exec.Command("fzf", "--filter="+query, "--delimiter=:", "--with-nth=1,2,3,4")
	fzf.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	fzf.Stdout = &out
	if err := fzf.Run(); err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(out.String()), "\n"), nil
}
