package fileexplorer

import (
	"nvim-gui/utils"

	"github.com/charmbracelet/log"
)

func (e *FileExplorer) onClose(_ ...any) {
	e.RequestClose()
}

func (e *FileExplorer) onCursorMoved(data ...any) {
	if len(data) == 0 {
		return
	}

	row, col := 0, 0
	switch cursor := data[0].(type) {
	case []uint64:
		if len(cursor) < 2 {
			return
		}
		row, col = int(cursor[0]), int(cursor[1])
	case []interface{}:
		if len(cursor) < 2 {
			return
		}
		row, col = utils.ReflectToInt(cursor[0]), utils.ReflectToInt(cursor[1])
	default:
		if len(data) < 2 {
			return
		}
		row, col = utils.ReflectToInt(data[0]), utils.ReflectToInt(data[1])
	}

	log.Infof("FileExplorerCursorMoved row=%d col=%d", row, col)
	if !e.GetActive() {
		return
	}
	if err := e.UpdateSelectedyEntry(row-1, col); err != nil {
		log.Error(err.Error())
	}
}

func (e *FileExplorer) onBufferLines(data ...any) {
	log.Infof("[nvim_buf_lines_event] raw data: %v", data)
	if !e.GetActive() {
		return
	}
	if len(data) != 6 {
		log.Info("[nvim_buf_lines_event] unexpected data length")
		return
	}

	buf := utils.ReflectToInt(data[0])
	firstline := utils.ReflectToInt(data[2])
	lastline := utils.ReflectToInt(data[3])

	var linedata []string
	switch lines := data[4].(type) {
	case []string:
		linedata = lines
	case []interface{}:
		linedata = utils.MapArray(lines, func(item interface{}) string {
			line, _ := item.(string)
			return line
		})
	default:
		log.Info("[nvim_buf_lines_event] couldnt parse data")
		return
	}

	if err := e.UpdateEntries(buf, firstline, lastline, linedata); err != nil {
		log.Errorf("FileExplorer update entry text failed: %v", err)
	}
}

func (e *FileExplorer) onBufferDetach(data ...any) {
	if len(data) == 0 {
		return
	}
	log.Info("nvim_buf_detach_event", "buf", utils.ReflectToInt(data[0]))
}

func (e *FileExplorer) onGoIn(_ ...any) {
	if err := e.GoIn(); err != nil {
		log.Errorf("FileExplorerGoIn failed: %v", err)
	}
}

func (e *FileExplorer) onGoOut(_ ...any) {
	if err := e.GoOut(); err != nil {
		log.Errorf("FileExplorerGoOut failed: %v", err)
	}
}

func (e *FileExplorer) onSync(_ ...any) {
	if err := e.Sync(); err != nil {
		log.Errorf("FileExplorerSync failed: %v", err)
	}
}
