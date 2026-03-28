package neovim

import (
	"nvim-gui/rendering"

	"github.com/charmbracelet/log"
)

func (s *Screen) optimizeGrid(grid *Grid, filetype string, bufNr, cursorLine int) []rendering.ContentRow {
	data := s.toRenderingGridData(grid)
	payload := rendering.BuildContentPayload(rendering.ContentInput{
		WindowID:   0,
		Filetype:   filetype,
		BufNr:      bufNr,
		CursorLine: cursorLine,
		Grid:       data,
		Meta:       s,
	})
	if grid != nil {
		dirtyCount := 0
		for _, dirty := range grid.DirtyRows {
			if dirty {
				dirtyCount++
			}
		}
		log.Debug("optimizeGrid payload", "grid_id", grid.ID, "filetype", filetype, "buf_nr", bufNr, "cursor_line", cursorLine, "rows_out", len(payload.Content), "dirty_rows", dirtyCount)
	}
	return payload.Content
}

func (s *Screen) HeadingLevel(bufNr, bufferLine int) int {
	s.headingMetadataMu.RLock()
	defer s.headingMetadataMu.RUnlock()
	headings, ok := s.headingMetadata[bufNr]
	if !ok {
		return 0
	}
	return headings[bufferLine]
}

func (s *Screen) TableMetaForLine(bufNr, bufferLine int) *rendering.TableMeta {
	s.tableMetadataMu.RLock()
	defer s.tableMetadataMu.RUnlock()
	tables := s.tableMetadata[bufNr]
	for i := range tables {
		t := &tables[i]
		if bufferLine >= t.StartLine && bufferLine < t.EndLine {
			return t
		}
	}
	return nil
}

func (s *Screen) ImageMetaForLine(bufNr, bufferLine int) *rendering.ImageMeta {
	s.imageMetadataMu.RLock()
	defer s.imageMetadataMu.RUnlock()
	images := s.imageMetadata[bufNr]
	for i := range images {
		im := &images[i]
		if im.Line == bufferLine {
			return im
		}
	}
	return nil
}

func (s *Screen) TaskMetaForLine(bufNr, bufferLine int) (checked bool, isTask bool) {
	s.taskMetadataMu.RLock()
	defer s.taskMetadataMu.RUnlock()
	tasks, ok := s.taskMetadata[bufNr]
	if !ok {
		return false, false
	}
	v, exists := tasks[bufferLine]
	return v, exists
}

func (s *Screen) CodeBlockMetaForLine(bufNr, bufferLine int) *rendering.CodeBlockMeta {
	s.codeBlockMetadataMu.RLock()
	defer s.codeBlockMetadataMu.RUnlock()
	blocks := s.codeBlockMetadata[bufNr]
	for i := range blocks {
		b := &blocks[i]
		if bufferLine >= b.StartLine && bufferLine < b.EndLine {
			return b
		}
	}
	return nil
}

func (s *Screen) InlineCodeRanges(bufNr, bufferLine int) []rendering.ColRange {
	s.inlineCodeMetadataMu.RLock()
	defer s.inlineCodeMetadataMu.RUnlock()
	lines, ok := s.inlineCodeMetadata[bufNr]
	if !ok {
		return nil
	}
	return lines[bufferLine]
}

func (s *Screen) toRenderingGridData(grid *Grid) *rendering.GridData {
	if grid == nil {
		return nil
	}
	return &rendering.GridData{
		Height:        grid.Height,
		TopLine:       grid.TopLine,
		DirtyRows:     grid.DirtyRows,
		Cells:         grid.Cells,
		OptimizedRows: grid.OptimizedRows,
		CachedTokens:  grid.CachedTokens,
		MarkdownOpts:  grid.MarkdownOpts,
		Cursor: rendering.CursorPosition{
			Row: grid.Cursor.Row,
			Col: grid.Cursor.Col,
		},
	}
}
