import { SubstituteJump } from '$lib/wailsjs/go/main/App';

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
