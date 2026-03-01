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
              local resolved = vim.fn.resolve(buf_dir .. '/' .. url)
              url = resolved
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

local group = vim.api.nvim_create_augroup('NvimGuiMarkdownTables', { clear = true })

vim.api.nvim_create_autocmd({ 'BufEnter', 'TextChanged', 'TextChangedI', 'InsertLeave' }, {
  group = group,
  pattern = '*.md',
  callback = function(ev)
    disable_markdown_renderers()
    send_tables(ev.buf)
    send_images(ev.buf)
    send_headings(ev.buf)
  end,
})

-- Run immediately for the current buffer (VimEnter/BufEnter already fired before this Lua loads)
disable_markdown_renderers()
send_tables(vim.api.nvim_get_current_buf())
send_images(vim.api.nvim_get_current_buf())
send_headings(vim.api.nvim_get_current_buf())
`
