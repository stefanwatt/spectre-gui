package neovim

import (
	"nvim-gui/utils"
)

// CompletionItem represents a single completion item sent to the frontend
type CompletionItem struct {
	Label      string `json:"label"`
	Kind       string `json:"kind"`   // "Function", "Variable", etc.
	Detail     string `json:"detail"` // type signature
	Source     string `json:"source"` // "lsp", "buffer", "path"
	Deprecated bool   `json:"deprecated"`
	SourceName string `json:"sourceName"` // "LSP", "Buffer", "Path"
}

// CompletionState represents the full completion menu state
type CompletionState struct {
	Items         []CompletionItem `json:"items"`
	SelectedIndex int              `json:"selectedIndex"` // -1 if none (using 1-indexed to match blink.cmp)
	Col           int              `json:"col"`           // cursor col for positioning
}

// CompletionDocumentation represents documentation for the selected item
type CompletionDocumentation struct {
	Text   string `json:"text"`   // documentation text
	Kind   string `json:"kind"`   // "markdown" or "plaintext"
	Detail string `json:"detail"` // item detail (type signature)
}

// LSP CompletionItemKind number -> name mapping
var kindNames = []string{
	"",              // 0 (unused)
	"Text",          // 1
	"Method",        // 2
	"Function",      // 3
	"Constructor",   // 4
	"Field",         // 5
	"Variable",      // 6
	"Class",         // 7
	"Interface",     // 8
	"Module",        // 9
	"Property",      // 10
	"Unit",          // 11
	"Value",         // 12
	"Enum",          // 13
	"Keyword",       // 14
	"Snippet",       // 15
	"Color",         // 16
	"File",          // 17
	"Reference",     // 18
	"Folder",        // 19
	"EnumMember",    // 20
	"Constant",      // 21
	"Struct",        // 22
	"Event",         // 23
	"Operator",      // 24
	"TypeParameter", // 25
}

func kindNumberToName(kind int) string {
	if kind >= 1 && kind < len(kindNames) {
		return kindNames[kind]
	}
	return "Text"
}

func parseCompletionItems(itemsRaw []interface{}) []CompletionItem {
	var items []CompletionItem
	for _, raw := range itemsRaw {
		entry, ok := raw.([]interface{})
		if !ok || len(entry) < 6 {
			continue
		}
		label, _ := entry[0].(string)
		kindNum := utils.ReflectToInt(entry[1])
		detail, _ := entry[2].(string)
		sourceId, _ := entry[3].(string)
		deprecated := utils.ReflectToInt(entry[4]) == 1
		sourceName, _ := entry[5].(string)

		items = append(items, CompletionItem{
			Label:      label,
			Kind:       kindNumberToName(kindNum),
			Detail:     detail,
			Source:     sourceId,
			Deprecated: deprecated,
			SourceName: sourceName,
		})
	}
	return items
}

// Embedded Lua bridge code for blink.cmp integration
const completionLua = `
local channel = ...

local function setup_completion_bridge()
  local ok, cmp = pcall(require, 'blink.cmp')
  if not ok then return false end

  local ok_list, list = pcall(require, 'blink.cmp.completion.list')
  if not ok_list then return false end

  -- Override is_menu_visible to check list state instead of window state
  -- This allows keymaps to work even when menu.enabled = false
  cmp.is_menu_visible = function()
    return #list.items > 0 and list.context ~= nil
  end

  -- Register listener for show events
  list.show_emitter:on(function(data)
    local items = list.items
    local selected_idx = list.selected_item_idx or -1
    local compact_items = {}

    for i, item in ipairs(items) do
      table.insert(compact_items, {
        item.label,
        item.kind or 0,
        item.detail or '',
        item.source_id or '',
        item.deprecated and 1 or 0,
        item.source_name or '',
      })
    end

    -- Get cursor col at trigger point for horizontal positioning
    local cursor_col = vim.api.nvim_win_get_cursor(0)[2]

    vim.rpcnotify(channel, 'CompletionShow', {compact_items, selected_idx, cursor_col})
  end)

  -- Register listener for hide events
  list.hide_emitter:on(function(data)
    vim.rpcnotify(channel, 'CompletionHide', {})
  end)

  -- Hook into documentation rendering to capture resolved documentation
  local function send_documentation()
    -- Get the currently selected item directly from the list
    local selected_item = list.get_selected_item()
    if not selected_item then
      vim.rpcnotify(channel, 'CompletionDocumentation', {'', 'plaintext', ''})
      return
    end

    -- Resolve the item to get documentation (if not already resolved)
    local ok_sources, sources = pcall(require, 'blink.cmp.sources.lib')
    if ok_sources and list.context then
      sources.resolve(list.context, selected_item):map(function(resolved_item)
        local doc_text = ''
        local doc_kind = 'plaintext'

        -- Extract documentation if available (after resolve)
        if resolved_item.documentation then
          if type(resolved_item.documentation) == 'string' then
            doc_text = resolved_item.documentation
            doc_kind = 'plaintext'
          elseif type(resolved_item.documentation) == 'table' then
            doc_text = resolved_item.documentation.value or ''
            doc_kind = resolved_item.documentation.kind or 'plaintext'
          end
        end

        local detail = resolved_item.detail or ''
        vim.rpcnotify(channel, 'CompletionDocumentation', {doc_text, doc_kind, detail})
      end)
    else
      -- Still send what we have without resolving
      local doc_text = ''
      local doc_kind = 'plaintext'

      if selected_item.documentation then
        if type(selected_item.documentation) == 'string' then
          doc_text = selected_item.documentation
        elseif type(selected_item.documentation) == 'table' then
          doc_text = selected_item.documentation.value or ''
          doc_kind = selected_item.documentation.kind or 'plaintext'
        end
      end

      local detail = selected_item.detail or ''
      vim.rpcnotify(channel, 'CompletionDocumentation', {doc_text, doc_kind, detail})
    end
  end

  -- Listen to selection changes and trigger documentation check
  list.select_emitter:on(function(data)
    vim.rpcnotify(channel, 'CompletionSelect', {data.idx or -1})

    -- Check documentation multiple times to catch async resolve
    vim.defer_fn(send_documentation, 50)
    vim.defer_fn(send_documentation, 200)
  end)

  return true
end

-- Try immediately (in case blink is already loaded)
if not setup_completion_bridge() then
  -- If not loaded yet, wait for first InsertEnter
  vim.api.nvim_create_autocmd('InsertEnter', {
    once = true,
    callback = function()
      -- Small delay to let blink.cmp initialize
      vim.defer_fn(function()
        setup_completion_bridge()
      end, 100)
    end,
  })
end
`
