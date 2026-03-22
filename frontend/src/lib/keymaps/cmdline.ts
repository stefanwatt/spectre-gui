import { SubstituteJump } from '@bindings/nvim-gui/app.js';

export const keymaps: App.Keymap[] = [
  {
    key: 'Tab',
    mode: 'cmdline',
    mods: [],
    action: (e) => {
      e.preventDefault()
      SubstituteJump()
    }
  },
]
