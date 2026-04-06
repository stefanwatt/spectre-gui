package features

// HoverDocumentation represents hover content received from LSP
type HoverDocumentation struct {
	Content string `json:"content"` // markdown content from LSP
	Row     int    `json:"row"`     // 1-based buffer line (cursor position)
	Col     int    `json:"col"`     // 0-based cursor col
}

// HoverLua sets up a K keymap that requests hover documentation directly
// from the LSP and sends the markdown content to Go via rpcnotify.
// No neovim floating window is created.
const HoverLua = `
local channel = ...
vim.keymap.set("n", "K", function()
  local params = vim.lsp.util.make_position_params()
  local cursor = vim.api.nvim_win_get_cursor(0)
  local responses = vim.lsp.buf_request_sync(0, "textDocument/hover", params, 2000)
  if not responses then
    return 
  end

  for _, resp in pairs(responses) do
    if resp.result and resp.result.contents then
      local contents = resp.result.contents
      local content = ""
      if type(contents) == "string" then
        content = contents
      elseif contents.value then
        content = contents.value
      end
      if content ~= "" then
        vim.rpcnotify(channel, "LspHover", {content, cursor[1], cursor[2]})
        vim.keymap.set("n", "q", function()
          vim.rpcnotify(channel, "LspHoverClose", {})
        end, { buffer = 0, once = true })
        return
      end
    end
  end
end)
`
