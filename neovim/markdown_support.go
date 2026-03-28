package neovim

import "nvim-gui/rendering"

const markdownTablesLua = `
return true
`

func parseMarkdownTables(tablesRaw []interface{}) []rendering.TableMeta {
	tables := make([]rendering.TableMeta, 0, len(tablesRaw))
	for _, raw := range tablesRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 3 {
			continue
		}
		meta := rendering.TableMeta{
			StartLine: toInt(entry[0]),
			EndLine:   toInt(entry[1]),
		}
		if aligns, ok := entry[2].([]interface{}); ok {
			for _, a := range aligns {
				if s, ok := a.(string); ok {
					meta.Alignments = append(meta.Alignments, s)
				}
			}
		}
		tables = append(tables, meta)
	}
	return tables
}

func parseMarkdownImages(imagesRaw []interface{}) []rendering.ImageMeta {
	images := make([]rendering.ImageMeta, 0, len(imagesRaw))
	for _, raw := range imagesRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 3 {
			continue
		}
		url, _ := entry[1].(string)
		alt, _ := entry[2].(string)
		images = append(images, rendering.ImageMeta{Line: toInt(entry[0]), URL: url, AltText: alt})
	}
	return images
}

func parseMarkdownHeadings(headingsRaw []interface{}) map[int]int {
	headings := make(map[int]int)
	for _, raw := range headingsRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		headings[toInt(entry[0])] = toInt(entry[1])
	}
	return headings
}

func parseMarkdownTasks(tasksRaw []interface{}) map[int]bool {
	tasks := make(map[int]bool)
	for _, raw := range tasksRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		checked, _ := entry[1].(bool)
		tasks[toInt(entry[0])] = checked
	}
	return tasks
}

func parseMarkdownCodeBlocks(blocksRaw []interface{}) []rendering.CodeBlockMeta {
	blocks := make([]rendering.CodeBlockMeta, 0, len(blocksRaw))
	for _, raw := range blocksRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		blocks = append(blocks, rendering.CodeBlockMeta{StartLine: toInt(entry[0]), EndLine: toInt(entry[1])})
	}
	return blocks
}

func parseMarkdownInlineCode(codesRaw []interface{}) map[int][]rendering.ColRange {
	codes := make(map[int][]rendering.ColRange)
	for _, raw := range codesRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		line := toInt(entry[0])
		rangesRaw, ok := entry[1].([]interface{})
		if !ok {
			continue
		}
		ranges := make([]rendering.ColRange, 0, len(rangesRaw))
		for _, rr := range rangesRaw {
			r, ok := rr.([]interface{})
			if !ok || len(r) < 2 {
				continue
			}
			ranges = append(ranges, rendering.ColRange{StartCol: toInt(r[0]), EndCol: toInt(r[1])})
		}
		codes[line] = ranges
	}
	return codes
}

func (s *Screen) setTableMetadata(bufNr int, tables []rendering.TableMeta) {
	s.tableMetadataMu.Lock()
	defer s.tableMetadataMu.Unlock()
	s.tableMetadata[bufNr] = tables
}

func (s *Screen) setImageMetadata(bufNr int, images []rendering.ImageMeta) {
	s.imageMetadataMu.Lock()
	defer s.imageMetadataMu.Unlock()
	s.imageMetadata[bufNr] = images
}

func (s *Screen) setHeadingMetadata(bufNr int, headings map[int]int) {
	s.headingMetadataMu.Lock()
	defer s.headingMetadataMu.Unlock()
	s.headingMetadata[bufNr] = headings
}

func (s *Screen) setTaskMetadata(bufNr int, tasks map[int]bool) {
	s.taskMetadataMu.Lock()
	defer s.taskMetadataMu.Unlock()
	s.taskMetadata[bufNr] = tasks
}

func (s *Screen) setCodeBlockMetadata(bufNr int, blocks []rendering.CodeBlockMeta) {
	s.codeBlockMetadataMu.Lock()
	defer s.codeBlockMetadataMu.Unlock()
	s.codeBlockMetadata[bufNr] = blocks
}

func (s *Screen) setInlineCodeMetadata(bufNr int, codes map[int][]rendering.ColRange) {
	s.inlineCodeMetadataMu.Lock()
	defer s.inlineCodeMetadataMu.Unlock()
	s.inlineCodeMetadata[bufNr] = codes
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case uint64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}
