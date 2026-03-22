# nvim-gui
this project is a gui for neovim written in wails using svelte a frontend framework.
# data-flow
a headless neovim instance is running in the background and sends redraw events (via msgpack rpc api) 
that represent sections of the terminal grid. this is then parsed, tokenized(for html rendering performance)
and sent to the frontend for rendering. 
