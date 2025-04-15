package picker

import (
	undo "nvim-gui/picker/undo"
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
