# Neovim-GUI
## Motivation
There are plenty of projects out there already implementing the neovim [ui protocol](https://neovim.io/doc/user/api-ui-events/#ui-protocol).
- [neovide](https://github.com/neovide/neovide)
- [neovim-qt](https://github.com/equalsraf/neovim-qt)
- [goneovim](https://github.com/akiyosi/goneovim)
So why do we need another one?
All of these GUIs (afaik) don't really go beyond rendering the redraw events other than
some cursor animations etc.
Presumably because neovim only provides UI updates on the basis of the grid.
But a GUI that just renders the grid anyway is not very interesting to me.
I want to make use of the UI capabilities that set apart a proper GUI from a TUI. 
## Features
### Done
#### Markdown rendering
- Variable font-size for headings
- Rich UI for 
    - tables
    - tasks
    - quotes
- image rendering

#### Rich UI Elements
##### ext_cmdline
Basically we can do what [noice](https://github.com/folke/noice.nvim) is doing, just better.
Rendering stuff like a custom popup for the substitute command with little indicators is trivial with the power of HTML/CSS,
but requires jumping through a lot of hoops on a grid.

### Planned
#### Windows
- leverage wails v3 window feature 
- run in server mode and spawn new windows as needed
- could even spawn a new window for floats/splits
  such that you can use window manager keymaps for window navigation
## Known Issues
## Compatibility
- dont use smear-cursor
