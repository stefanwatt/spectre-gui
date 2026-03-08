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
#### Dashboard
#### File-Explorer
#### Windows
- leverage wails v3 window feature 
- run in server mode and spawn new windows as needed
- could even spawn a new window for floats/splits
  such that you can use window manager keymaps for window navigation

## Goals
I love neovim. It's great at almost everything. This project is there to help neovim out in areas where it's not so great.
### Rich UI
Rendering on a grid is simply a massive limitation for displaying rich UI components.
I'm not a fan of ricing and esthetics are not my main concern
I don't use a buffer-/tabline. 
I don't use a minimap.
I don't use a filetree.
UI is not just about looking fancy though.
#### Examples
- search and replace pop-up with smaller font and little indicators for flags
- status line can have smaller sub-elements

### Windows
#### Floating
Rendering floating windows on a terminal grid is a bit ridiculous. 
If you want a border around a 3x3 window, well then now your window is 4x4.
This project strives to render any floating windows via custom HTML rather than rendering the grid.

Floating windows anchored to the global editor (e.g. hover docs, signature help) are spawned as 
separate frameless OS windows with the title/name `nvim-gui-float`. To make your window manager 
treat them as floating, add the appropriate rule:

```
# Hyprland
windowrulev2 = float, title:^(nvim-gui-float)$

# Sway
for_window [title="nvim-gui-float"] floating enable

# i3
for_window [title="nvim-gui-float"] floating enable
```
#### Tiled
I love my tiling window manager and I want it to be in charge of my window management. 
I don't want to introduce another set of keymaps to manage my neovim windows.
Solution seems obvious: Every neovim window is an OS window (except floating).
## Non-Goals
### Performance
This thing will never be as performant as Neovide or Zed.
The web-stack used is simply not able to achieve that.
That being said wails apps are still plenty fast for me.
## Known Issues
- window size does not span the full os window. resizing os window will make it work.
- non-existent lines are being rendered to fill the rest of the available space
### Completion
- command line completion window is rendered in the wrong spot
  needs to be anchored to the custom svelte input
### Markdown
- sometimes heading sizing is applied to the wrong line or not applied to lines where it should be 
  moving the cursor across the line or open+close mini.files fixes it -> suggests desynched state 
- multiline quotes are broken: causes the second line to be rendered in two places
### Highlighting
- cursor movement in window a causes highlighting changes in window b
## Compatibility
- dont use smear-cursor
