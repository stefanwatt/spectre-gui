package rendering

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
)

type Renderer struct{
	PendingRender        bool
}

func (r *Renderer) ScheduleRender() {
	if !r.PendingRender {
		r.PendingRender = true
		// Use goroutine to simulate setImmediate behavior
		go func() {
			// Small delay to batch updates
			time.Sleep(time.Millisecond * 5)
			r.render()
			r.PendingRender = false
		}()
	}
}

func (r *Renderer) render() {
	r.EmitFloatingWindows()
	r.updateLayout()

	hasFileExplorerDirty := false
	for winId, window := range r.Windows {
		if window.Hidden {
			continue
		}
		if !window.Dirty {
			continue
		}
		window.Dirty = false
		grid := window.Grid

		if r.FileExplorer != nil && window.IsFileExplorer {
			bufNr := window.Buffer.BufNr
			log.Debug(fmt.Sprintf("[minifiles] render: processing file explorer winId=%d parentWinId=%d currentWinId=%d",
				winId, r.FileExplorer.Parent.WinId, r.FileExplorer.Current.WinId))
			if r.FileExplorer.Parent.WinId == winId {
				log.Debug(fmt.Sprintf("[minifiles] render: updating parent bufNr=%d", bufNr))
				r.FileExplorer.updateParent(grid, bufNr)
			} else if r.FileExplorer.Current.WinId == winId {
				log.Debug(fmt.Sprintf("[minifiles] render: updating current bufNr=%d", bufNr))
				r.FileExplorer.updateCurrent(grid, bufNr)
			} else if r.FileExplorer.Preview != nil && r.FileExplorer.Preview.GetWinId() == winId {
				// Determine preview type from the currently selected entry in the current directory
				selectedIsDir := false
				for _, entry := range r.FileExplorer.Current.Entries {
					if entry.ID == r.FileExplorer.Current.SelectedEntryId {
						selectedIsDir = entry.IsDir
						break
					}
				}
				if selectedIsDir {
					log.Debug(fmt.Sprintf("[minifiles] render: updating dir preview bufNr=%d", bufNr))
					r.FileExplorer.updateDirPreview(grid, bufNr)
				} else {
					log.Debug(fmt.Sprintf("[minifiles] render: updating content preview bufNr=%d", bufNr))
					r.FileExplorer.updateContentPreview(r.optimizeGrid(grid, "minifiles", bufNr, -1))
				}
			}
			hasFileExplorerDirty = true
			continue
		}

		if !window.IsFloating() || window.ZIndex == 69420 {
			filetype := ""
			bufNr := 0
			if window.Buffer != nil {
				filetype = (*window.Buffer).Filetype
				bufNr = (*window.Buffer).BufNr
			}
			if filetype == "markdown" {
				log.Debug(fmt.Sprintf("render: winId=%d filetype=%s bufNr=%d", winId, filetype, bufNr))
			}
			cursorLine := -1
			if window.Cursor != nil {
				cursorLine = window.Cursor.Row - 1 // convert 1-indexed to 0-indexed
			}
			r.emitEvent("content-updated", map[string]interface{}{
				"winId":          winId,
				"updatedContent": r.optimizeGrid(grid, filetype, bufNr, cursorLine),
			})
		} else {
			// Skip rendering blink-cmp floating windows (handled by native completion menu)
			if window.Buffer != nil && (*window.Buffer).Filetype == "blink-cmp-menu" {
				continue
			}
			if r.app != nil {
				r.app.Event.Emit("content-updated", map[string]interface{}{
					"winId":          winId,
					"updatedContent": r.renderFloatingWindow(window),
				})
			}
		}
	}

	if hasFileExplorerDirty {
		log.Debug("[minifiles] render: emitting file-explorer-update")
		r.syncFileExplorerMode()
		r.emitEvent("file-explorer-update", r.FileExplorer)
		log.Debug("[minifiles] render: file-explorer-update emitted successfully")
	}
}
