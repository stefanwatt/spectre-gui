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
##### External Commandline
Basically we can do what [noice](https://github.com/folke/noice.nvim) is doing, just better.
Rendering stuff like a custom popup for the substitute command with little indicators is trivial with the power of HTML/CSS,
but requires jumping through a lot of hoops on a grid.

##### Native Completion Menu
Integrates with [blink.cmp](https://github.com/Saghen/blink.cmp) to provide a native Svelte completion menu with rich styling:
- Kind-specific icons and colors
- Deprecated item styling
- Smart positioning (flips above cursor when near bottom)
- Source indicators (LSP, Buffer, Path, etc.)

See [COMPLETION_SETUP.md](./COMPLETION_SETUP.md) for configuration instructions.

### Planned
#### Windows
- leverage wails v3 window feature 
- run in server mode and spawn new windows as needed
- could even spawn a new window for floats/splits
  such that you can use window manager keymaps for window navigation

## Non-Goals
### Performance
This thing will never be as performant as Neovide or Zed.
The web-stack used is simply not able to achieve that.
That being said wails apps are still plenty fast for me.
## Known Issues
## Compatibility
- dont use smear-cursor
