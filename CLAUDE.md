# Description
neovim gui
# Tech-Stack
## wails
- the framework for desktop apps
## go
- prefer to handle logic in go
- tokenize grid sent from neovim for more performant html rendering
## svelte
- svelte 5 runes are important for rendering performance and precise reactivity
- keep frontend as dumb as possible
- generally just render tokens, but leverage rich UI for picker and markdown rendering
## lua
some stuff must be handled in underlying neovim instance in lua
## nix
- provide some dependencies etc. in shell.nix -> run stuff like playwright tests in nix-shell

# features
## markdown 
neovim is bound to grid based rendering since its a TUI, 
but we can render variable font-size, images, etc.
to keep continuity while editing markdown the rich rendering is disabled if the cursor is on the table, image, etc.

# plans
- plan files go to `./docs/plans/`
- when a plan is finished you may put it into `./docs/plans/archive/`
