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
