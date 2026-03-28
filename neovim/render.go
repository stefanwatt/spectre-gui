package neovim

import (
	"fmt"

	"github.com/charmbracelet/log"
)

func (s *Screen) Render() {
	if s == nil {
		return
	}
	log.Debug("render start", "windows_total", len(s.Windows), "grids_total", len(s.Grids), "active_window", s.ActiveWindow)

	s.EmitFloatingWindows()
	s.updateLayout()

	hasFileExplorerDirty := false

	s.windowsMu.RLock()
	defer s.windowsMu.RUnlock()

	emittedContent := 0
	skippedHidden := 0
	skippedClean := 0
	emittedFloating := 0
	for winID, window := range s.Windows {
		if window == nil || window.Hidden || !window.Dirty {
			if window != nil && window.Hidden {
				skippedHidden++
			} else {
				skippedClean++
			}
			continue
		}
		log.Debug("render window dirty", "win_id", winID, "floating", window.IsFloating(), "zindex", window.ZIndex)
		window.Dirty = false
		grid := window.Grid
		if grid == nil {
			log.Debug("render skip nil grid", "win_id", winID)
			continue
		}

		if s.FileExplorer != nil && window.IsFileExplorer {
			bufNr := 0
			if window.Buffer != nil {
				bufNr = window.Buffer.BufNr
			}
			if s.FileExplorer.Parent.WinId == winID {
				s.FileExplorer.updateParent(grid, bufNr)
			} else if s.FileExplorer.Current.WinId == winID {
				s.FileExplorer.updateCurrent(grid, bufNr)
			} else if s.FileExplorer.Preview != nil && s.FileExplorer.Preview.GetWinId() == winID {
				selectedIsDir := false
				for _, entry := range s.FileExplorer.Current.Entries {
					if entry.ID == s.FileExplorer.Current.SelectedEntryId {
						selectedIsDir = entry.IsDir
						break
					}
				}
				if selectedIsDir {
					s.FileExplorer.updateDirPreview(grid, bufNr)
				} else {
					s.FileExplorer.updateContentPreview(s.optimizeGrid(grid, "minifiles", bufNr, -1))
				}
			}
			hasFileExplorerDirty = true
			continue
		}

		if !window.IsFloating() || window.ZIndex == 69420 {
			filetype := ""
			bufNr := 0
			if window.Buffer != nil {
				filetype = window.Buffer.Filetype
				bufNr = window.Buffer.BufNr
			}
			cursorLine := -1
			if window.Cursor != nil {
				cursorLine = window.Cursor.Row - 1
			}
			EmitEvent("content-updated", map[string]interface{}{
				"winId":          winID,
				"updatedContent": s.optimizeGrid(grid, filetype, bufNr, cursorLine),
			})
			emittedContent++
		} else {
			if window.Buffer != nil && window.Buffer.Filetype == "blink-cmp-menu" {
				continue
			}
			EmitEvent("content-updated", map[string]interface{}{
				"winId":          winID,
				"updatedContent": s.renderFloatingWindow(window),
			})
			emittedFloating++
		}
	}

	if hasFileExplorerDirty {
		log.Debug("[minifiles] render: emitting file-explorer-update")
		s.syncFileExplorerMode()
		EmitEvent("file-explorer-update", s.FileExplorer)
	}

	log.Debug(fmt.Sprintf("render cycle complete windows=%d", len(s.Windows)), "emitted_content", emittedContent, "emitted_floating", emittedFloating, "skipped_hidden", skippedHidden, "skipped_clean", skippedClean)
}
