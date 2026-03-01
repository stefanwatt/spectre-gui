package neovim

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

local function send_tables(bufnr)
  if vim.bo[bufnr].filetype ~= 'markdown' then
    return
  end
  local tables = get_table_metadata(bufnr)
  vim.rpcnotify(channel, 'MarkdownTables', {bufnr, tables})
end

local group = vim.api.nvim_create_augroup('NvimGuiMarkdownTables', { clear = true })

vim.api.nvim_create_autocmd({ 'BufEnter', 'TextChanged', 'TextChangedI', 'InsertLeave' }, {
  group = group,
  pattern = '*.md',
  callback = function(ev)
    disable_markdown_renderers()
    send_tables(ev.buf)
  end,
})

-- Run immediately for the current buffer (VimEnter/BufEnter already fired before this Lua loads)
disable_markdown_renderers()
send_tables(vim.api.nvim_get_current_buf())
`
