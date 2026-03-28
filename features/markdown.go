package features

import (
	"encoding/base64"
	"nvim-gui/utils"
	"strings"
	"sync"
)

type Idk struct {
	tableMetadata        map[int][]TableMeta // bufNr -> []TableMeta
	tableMetadataMu      sync.RWMutex
	imageMetadata        map[int][]ImageMeta // bufNr -> []ImageMeta
	imageMetadataMu      sync.RWMutex
	headingMetadata      map[int]map[int]int // bufNr -> line -> heading level
	headingMetadataMu    sync.RWMutex
	taskMetadata         map[int]map[int]bool // bufNr -> line -> checked
	taskMetadataMu       sync.RWMutex
	codeBlockMetadata    map[int][]CodeBlockMeta // bufNr -> []CodeBlockMeta
	codeBlockMetadataMu  sync.RWMutex
	inlineCodeMetadata   map[int]map[int][]ColRange // bufNr -> line -> []ColRange
	inlineCodeMetadataMu sync.RWMutex
}

//   tableMetadata:      make(map[int][]TableMeta),
// imageMetadata:      make(map[int][]ImageMeta),
// headingMetadata:    make(map[int]map[int]int),
// taskMetadata:       make(map[int]map[int]bool),
// codeBlockMetadata:  make(map[int][]CodeBlockMeta),
// inlineCodeMetadata: make(map[int]map[int][]ColRange),

type MarkdownOpts struct {
	QuoteLevel   int            `json:"quoteLevel"`
	HeadingLevel int            `json:"headingLevel,omitempty"`
	Table        *TableRowOpts  `json:"table,omitempty"`
	Image        *ImageOpts     `json:"image,omitempty"`
	Task         *TaskOpts      `json:"task,omitempty"`
	CodeBlock    *CodeBlockOpts `json:"codeBlock,omitempty"`
}

type TableMeta struct {
	StartLine  int      // 0-indexed buffer line
	EndLine    int      // exclusive
	Alignments []string // "left", "center", "right"
}

type TableRowOpts struct {
	TableID    int        `json:"tableId"`
	RowType    string     `json:"rowType"` // "header", "separator", "data"
	Cells      [][]*Token `json:"cells"`
	Alignments []string   `json:"alignments"`
}

type ImageMeta struct {
	Line    int // 0-indexed buffer line
	URL     string
	AltText string
}

type ImageOpts struct {
	URL     string `json:"url"`
	AltText string `json:"altText"`
}

type TaskOpts struct {
	Checked bool `json:"checked"`
}

type CodeBlockMeta struct {
	StartLine int // 0-indexed buffer line (inclusive)
	EndLine   int // 0-indexed buffer line (exclusive)
}

func buildTableRowOpts(meta *TableMeta, bufferLine int, tokens []*Token) *TableRowOpts {
	// Detect row type based on content, not just position
	// Box-drawing renderers may add border rows and change separators
	if isBorderRow(tokens) {
		return &TableRowOpts{
			TableID:    meta.StartLine,
			RowType:    "separator",
			Cells:      nil,
			Alignments: meta.Alignments,
		}
	}
	if isSeparatorRow(tokens) {
		return &TableRowOpts{
			TableID:    meta.StartLine,
			RowType:    "separator",
			Cells:      nil,
			Alignments: meta.Alignments,
		}
	}

	cells := splitTokensIntoCells(tokens)
	rowType := "data"
	if bufferLine == meta.StartLine {
		rowType = "header"
	}
	return &TableRowOpts{
		TableID:    meta.StartLine,
		RowType:    rowType,
		Cells:      cells,
		Alignments: meta.Alignments,
	}
}

func NewMarkdownOpts() *MarkdownOpts {
	return &MarkdownOpts{QuoteLevel: 0}
}

const markdownTablesLua = `
local channel = ...

-- Disable markdown rendering plugins that replace table characters
local function disable_markdown_renderers()
  pcall(function()
    if package.loaded['markview'] then
      vim.cmd('Markview disable')
    end
  end)
  pcall(function()
    if package.loaded['render-markdown'] then
      vim.cmd('RenderMarkdown disable')
    end
  end)
end

local function get_table_metadata(bufnr)
  local ok, parser = pcall(vim.treesitter.get_parser, bufnr, 'markdown')
  if not ok or not parser then
    return {}
  end

  local trees = parser:parse()
  if not trees or #trees == 0 then
    return {}
  end

  local root = trees[1]:root()
  local tables = {}

  -- Recursively walk the tree to find pipe_table nodes (they nest inside section nodes)
  local function walk(node)
    if node:type() == 'pipe_table' then
      local start_row, _, end_row, _ = node:range()

      -- Parse alignments from delimiter row
      local alignments = {}
      for table_child in node:iter_children() do
        if table_child:type() == 'pipe_table_delimiter_row' then
          local delim_text = vim.treesitter.get_node_text(table_child, bufnr)
          for cell in delim_text:gmatch('[^|]+') do
            cell = vim.trim(cell)
            if cell ~= '' then
              local left = cell:sub(1, 1) == ':'
              local right = cell:sub(-1) == ':'
              if left and right then
                table.insert(alignments, 'center')
              elseif right then
                table.insert(alignments, 'right')
              else
                table.insert(alignments, 'left')
              end
            end
          end
          break
        end
      end

      table.insert(tables, { start_row, end_row, alignments })
    else
      for child in node:iter_children() do
        walk(child)
      end
    end
  end

  walk(root)
  return tables
end

local function get_heading_metadata(bufnr)
  local ok, parser = pcall(vim.treesitter.get_parser, bufnr, 'markdown')
  if not ok or not parser then
    return {}
  end

  local trees = parser:parse()
  if not trees or #trees == 0 then
    return {}
  end

  local root = trees[1]:root()
  local headings = {}

  local function walk(node)
    local ntype = node:type()
    if ntype == 'atx_heading' then
      local start_row = node:range()
      -- Determine level from the marker child
      for child in node:iter_children() do
        local ctype = child:type()
        if ctype:match('^atx_h%d_marker$') then
          local level = tonumber(ctype:match('%d'))
          table.insert(headings, { start_row, level })
          break
        end
      end
    else
      for child in node:iter_children() do
        walk(child)
      end
    end
  end

  walk(root)
  return headings
end

local function get_image_metadata(bufnr)
  local ok, parser = pcall(vim.treesitter.get_parser, bufnr, 'markdown_inline')
  if not ok or not parser then
    return {}
  end

  local trees = parser:parse()
  if not trees or #trees == 0 then
    return {}
  end

  local images = {}
  local buf_dir = vim.fn.fnamemodify(vim.api.nvim_buf_get_name(bufnr), ':p:h')

  for _, tree in ipairs(trees) do
    local root = tree:root()
    local function walk(node)
      if node:type() == 'image' then
        local start_row, start_col, end_row, end_col = node:range()
        -- Only standalone images: must be on a single line and span the full line content
        if start_row == end_row then
          local line_text = vim.api.nvim_buf_get_lines(bufnr, start_row, start_row + 1, false)[1] or ''
          local trimmed = vim.trim(line_text)
          local node_text = vim.treesitter.get_node_text(node, bufnr)
          if trimmed == node_text then
            local url = ''
            local alt = ''
            for child in node:iter_children() do
              if child:type() == 'image_description' then
                alt = vim.treesitter.get_node_text(child, bufnr)
              elseif child:type() == 'link_destination' then
                url = vim.treesitter.get_node_text(child, bufnr)
              end
            end
            -- Resolve relative local paths
            if url ~= '' and not url:match('^https?://') then
              if url:sub(1, 2) == '~/' then
                url = vim.fn.expand('~') .. url:sub(2)
              elseif url:sub(1, 1) ~= '/' then
                url = buf_dir .. '/' .. url
              end
              url = vim.fn.resolve(url)
            end
            if url ~= '' then
              table.insert(images, { start_row, url, alt })
            end
          end
        end
      else
        for child in node:iter_children() do
          walk(child)
        end
      end
    end
    walk(root)
  end

  return images
end

local function send_tables(bufnr)
  if vim.bo[bufnr].filetype ~= 'markdown' then
    return
  end
  local tables = get_table_metadata(bufnr)
  vim.rpcnotify(channel, 'MarkdownTables', {bufnr, tables})
end

local function send_images(bufnr)
  if vim.bo[bufnr].filetype ~= 'markdown' then
    return
  end
  local images = get_image_metadata(bufnr)
  vim.rpcnotify(channel, 'MarkdownImages', {bufnr, images})
end

local function send_headings(bufnr)
  if vim.bo[bufnr].filetype ~= 'markdown' then
    return
  end
  local headings = get_heading_metadata(bufnr)
  vim.rpcnotify(channel, 'MarkdownHeadings', {bufnr, headings})
end

local function get_task_metadata(bufnr)
  local ok, parser = pcall(vim.treesitter.get_parser, bufnr, 'markdown')
  if not ok or not parser then
    return {}
  end

  local trees = parser:parse()
  if not trees or #trees == 0 then
    return {}
  end

  local root = trees[1]:root()
  local tasks = {}

  local function walk(node)
    local ntype = node:type()
    if ntype == 'task_list_marker_checked' or ntype == 'task_list_marker_unchecked' then
      local start_row = node:range()
      local checked = ntype == 'task_list_marker_checked' and 1 or 0
      table.insert(tasks, { start_row, checked })
    else
      for child in node:iter_children() do
        walk(child)
      end
    end
  end

  walk(root)
  return tasks
end

local function send_tasks(bufnr)
  if vim.bo[bufnr].filetype ~= 'markdown' then
    return
  end
  local tasks = get_task_metadata(bufnr)
  vim.rpcnotify(channel, 'MarkdownTasks', {bufnr, tasks})
end

local function get_codeblock_metadata(bufnr)
  local ok, parser = pcall(vim.treesitter.get_parser, bufnr, 'markdown')
  if not ok or not parser then
    return {}
  end

  local trees = parser:parse()
  if not trees or #trees == 0 then
    return {}
  end

  local root = trees[1]:root()
  local codeblocks = {}

  local function walk(node)
    if node:type() == 'fenced_code_block' then
      local start_row, _, end_row, _ = node:range()
      table.insert(codeblocks, { start_row, end_row })
    else
      for child in node:iter_children() do
        walk(child)
      end
    end
  end

  walk(root)
  return codeblocks
end

local function get_inline_code_metadata(bufnr)
  local ok, parser = pcall(vim.treesitter.get_parser, bufnr, 'markdown_inline')
  if not ok or not parser then
    return {}
  end

  local trees = parser:parse()
  if not trees or #trees == 0 then
    return {}
  end

  local results = {}
  for _, tree in ipairs(trees) do
    local root = tree:root()
    local function walk(node)
      if node:type() == 'code_span' then
        local start_row, start_col, end_row, end_col = node:range()
        -- Only handle single-line code spans
        if start_row == end_row then
          table.insert(results, { start_row, start_col, end_col })
        end
      else
        for child in node:iter_children() do
          walk(child)
        end
      end
    end
    walk(root)
  end

  return results
end

local function send_codeblocks(bufnr)
  if vim.bo[bufnr].filetype ~= 'markdown' then
    return
  end
  local codeblocks = get_codeblock_metadata(bufnr)
  vim.rpcnotify(channel, 'MarkdownCodeBlocks', {bufnr, codeblocks})
end

local function send_inline_code(bufnr)
  if vim.bo[bufnr].filetype ~= 'markdown' then
    return
  end
  local inline_codes = get_inline_code_metadata(bufnr)
  vim.rpcnotify(channel, 'MarkdownInlineCode', {bufnr, inline_codes})
end

local group = vim.api.nvim_create_augroup('NvimGuiMarkdownTables', { clear = true })

local debounce_timers = {}
local function send_all(bufnr)
  disable_markdown_renderers()
  send_tables(bufnr)
  send_images(bufnr)
  send_headings(bufnr)
  send_tasks(bufnr)
  send_codeblocks(bufnr)
  send_inline_code(bufnr)
end

-- BufEnter and InsertLeave fire immediately (structural changes on mode switch or file open)
vim.api.nvim_create_autocmd({ 'BufEnter', 'InsertLeave' }, {
  group = group,
  pattern = '*.md',
  callback = function(ev)
    local t = debounce_timers[ev.buf]
    if t then t:stop(); t:close(); debounce_timers[ev.buf] = nil end
    send_all(ev.buf)
  end,
})

-- TextChanged/TextChangedI are debounced: only parse after 250ms of no typing
vim.api.nvim_create_autocmd({ 'TextChanged', 'TextChangedI' }, {
  group = group,
  pattern = '*.md',
  callback = function(ev)
    local bufnr = ev.buf
    local t = debounce_timers[bufnr]
    if t then t:stop(); pcall(function() t:close() end) end
    local timer = vim.loop.new_timer()
    debounce_timers[bufnr] = timer
    timer:start(250, 0, vim.schedule_wrap(function()
      -- Guard: another keystroke may have replaced or already closed this timer
      if debounce_timers[bufnr] ~= timer then return end
      debounce_timers[bufnr] = nil
      pcall(function() timer:close() end)
      send_all(bufnr)
    end))
  end,
})

-- Run immediately for the current buffer (VimEnter/BufEnter already fired before this Lua loads)
disable_markdown_renderers()
send_tables(vim.api.nvim_get_current_buf())
send_images(vim.api.nvim_get_current_buf())
send_headings(vim.api.nvim_get_current_buf())
send_tasks(vim.api.nvim_get_current_buf())
send_codeblocks(vim.api.nvim_get_current_buf())
send_inline_code(vim.api.nvim_get_current_buf())
`

func (s *Screen) setTableMetadata(bufNr int, tables []TableMeta) {
	s.tableMetadataMu.Lock()
	defer s.tableMetadataMu.Unlock()
	s.tableMetadata[bufNr] = tables
}

func (s *Screen) getTableMetaForLine(bufNr, bufferLine int) *TableMeta {
	if bufNr == 0 {
		return nil
	}
	s.tableMetadataMu.RLock()
	defer s.tableMetadataMu.RUnlock()
	tables, exists := s.tableMetadata[bufNr]
	if !exists {
		return nil
	}
	for i := range tables {
		// Extend range by 1 on each side to capture box-drawing border rows
		// added by render-markdown plugins (e.g., ┌─┬─┐ and └─┴─┘)
		if bufferLine >= tables[i].StartLine-1 && bufferLine < tables[i].EndLine+1 {
			return &tables[i]
		}
	}
	return nil
}

func parseMarkdownTables(tablesRaw []interface{}) []TableMeta {
	var tables []TableMeta
	for _, raw := range tablesRaw {
		tableData, ok := raw.([]interface{})
		if !ok || len(tableData) < 3 {
			continue
		}
		startLine := utils.ReflectToInt(tableData[0])
		endLine := utils.ReflectToInt(tableData[1])
		alignmentsRaw, ok := tableData[2].([]interface{})
		if !ok {
			continue
		}
		alignments := make([]string, len(alignmentsRaw))
		for i, a := range alignmentsRaw {
			if s, ok := a.(string); ok {
				alignments[i] = s
			} else {
				alignments[i] = "left"
			}
		}
		tables = append(tables, TableMeta{
			StartLine:  startLine,
			EndLine:    endLine,
			Alignments: alignments,
		})
	}
	return tables
}

func (s *Screen) setImageMetadata(bufNr int, images []ImageMeta) {
	s.imageMetadataMu.Lock()
	defer s.imageMetadataMu.Unlock()
	s.imageMetadata[bufNr] = images
}

func (s *Screen) getImageMetaForLine(bufNr, bufferLine int) *ImageMeta {
	if bufNr == 0 {
		return nil
	}
	s.imageMetadataMu.RLock()
	defer s.imageMetadataMu.RUnlock()
	images, exists := s.imageMetadata[bufNr]
	if !exists {
		return nil
	}
	for i := range images {
		if images[i].Line == bufferLine {
			return &images[i]
		}
	}
	return nil
}

func parseMarkdownImages(imagesRaw []interface{}) []ImageMeta {
	var images []ImageMeta
	for _, raw := range imagesRaw {
		imageData, ok := raw.([]interface{})
		if !ok || len(imageData) < 3 {
			continue
		}
		line := utils.ReflectToInt(imageData[0])
		url, ok := imageData[1].(string)
		if !ok {
			continue
		}
		alt, _ := imageData[2].(string)

		// Rewrite local paths to use the local-image server
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			encoded := base64.URLEncoding.EncodeToString([]byte(url))
			url = "/local-image/" + encoded
		}

		images = append(images, ImageMeta{
			Line:    line,
			URL:     url,
			AltText: alt,
		})
	}
	return images
}

func (s *Screen) setHeadingMetadata(bufNr int, headings map[int]int) {
	s.headingMetadataMu.Lock()
	defer s.headingMetadataMu.Unlock()
	s.headingMetadata[bufNr] = headings
}

func (s *Screen) getHeadingLevel(bufNr, bufferLine int) int {
	if bufNr == 0 {
		return 0
	}
	s.headingMetadataMu.RLock()
	defer s.headingMetadataMu.RUnlock()
	headings, exists := s.headingMetadata[bufNr]
	if !exists {
		return 0
	}
	return headings[bufferLine]
}

func parseMarkdownHeadings(headingsRaw []interface{}) map[int]int {
	headings := make(map[int]int)
	for _, raw := range headingsRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		line := utils.ReflectToInt(entry[0])
		level := utils.ReflectToInt(entry[1])
		headings[line] = level
	}
	return headings
}

func (s *Screen) setTaskMetadata(bufNr int, tasks map[int]bool) {
	s.taskMetadataMu.Lock()
	defer s.taskMetadataMu.Unlock()
	s.taskMetadata[bufNr] = tasks
}

func (s *Screen) getTaskMetaForLine(bufNr, bufferLine int) (checked, isTask bool) {
	if bufNr == 0 {
		return false, false
	}
	s.taskMetadataMu.RLock()
	defer s.taskMetadataMu.RUnlock()
	tasks, exists := s.taskMetadata[bufNr]
	if !exists {
		return false, false
	}
	checked, isTask = tasks[bufferLine]
	return checked, isTask
}

func parseMarkdownTasks(tasksRaw []interface{}) map[int]bool {
	tasks := make(map[int]bool)
	for _, raw := range tasksRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		line := utils.ReflectToInt(entry[0])
		checked := utils.ReflectToInt(entry[1]) == 1
		tasks[line] = checked
	}
	return tasks
}

func (s *Screen) setCodeBlockMetadata(bufNr int, blocks []CodeBlockMeta) {
	s.codeBlockMetadataMu.Lock()
	defer s.codeBlockMetadataMu.Unlock()
	s.codeBlockMetadata[bufNr] = blocks
}

func (s *Screen) getCodeBlockMetaForLine(bufNr, bufferLine int) *CodeBlockMeta {
	if bufNr == 0 {
		return nil
	}
	s.codeBlockMetadataMu.RLock()
	defer s.codeBlockMetadataMu.RUnlock()
	blocks, exists := s.codeBlockMetadata[bufNr]
	if !exists {
		return nil
	}
	for i := range blocks {
		if bufferLine >= blocks[i].StartLine && bufferLine < blocks[i].EndLine {
			return &blocks[i]
		}
	}
	return nil
}

func parseMarkdownCodeBlocks(blocksRaw []interface{}) []CodeBlockMeta {
	var blocks []CodeBlockMeta
	for _, raw := range blocksRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		startLine := utils.ReflectToInt(entry[0])
		endLine := utils.ReflectToInt(entry[1])
		blocks = append(blocks, CodeBlockMeta{
			StartLine: startLine,
			EndLine:   endLine,
		})
	}
	return blocks
}

func (s *Screen) setInlineCodeMetadata(bufNr int, codes map[int][]ColRange) {
	s.inlineCodeMetadataMu.Lock()
	defer s.inlineCodeMetadataMu.Unlock()
	s.inlineCodeMetadata[bufNr] = codes
}

func (s *Screen) getInlineCodeRanges(bufNr, bufferLine int) []ColRange {
	if bufNr == 0 {
		return nil
	}
	s.inlineCodeMetadataMu.RLock()
	defer s.inlineCodeMetadataMu.RUnlock()
	lines, exists := s.inlineCodeMetadata[bufNr]
	if !exists {
		return nil
	}
	return lines[bufferLine]
}

func parseMarkdownInlineCode(codesRaw []interface{}) map[int][]ColRange {
	codes := make(map[int][]ColRange)
	for _, raw := range codesRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 3 {
			continue
		}
		line := utils.ReflectToInt(entry[0])
		startCol := utils.ReflectToInt(entry[1])
		endCol := utils.ReflectToInt(entry[2])
		codes[line] = append(codes[line], ColRange{StartCol: startCol, EndCol: endCol})
	}
	return codes
}
