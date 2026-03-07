package neovim

const externalSplitsLua = `
local function external_split(cmd_args)
  local file = cmd_args.args
  local buf
  if file and file ~= '' then
    buf = vim.fn.bufadd(file)
    vim.fn.bufload(buf)
  else
    buf = vim.api.nvim_get_current_buf()
  end
  vim.api.nvim_open_win(buf, true, {
    external = true,
    width = 80,
    height = 24
  })
end

vim.api.nvim_create_user_command('Split', external_split, { nargs = '?', complete = 'file' })
vim.api.nvim_create_user_command('VSplit', external_split, { nargs = '?', complete = 'file' })
vim.cmd('cabbrev split Split')
vim.cmd('cabbrev vsplit VSplit')
vim.cmd('cabbrev sp Split')
vim.cmd('cabbrev vsp VSplit')

vim.keymap.set('n', '<C-w>s', function()
  vim.api.nvim_open_win(0, true, {
    external = true,
    width = 80,
    height = 24
  })
end)
vim.keymap.set('n', '<C-w>v', function()
  vim.api.nvim_open_win(0, true, {
    external = true,
    width = 80,
    height = 24
  })
end)
`
