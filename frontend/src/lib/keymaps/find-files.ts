import { pickers } from "$lib/state.svelte";

export const keymaps: App.Keymap[]=[
  {
    key: 'Escape',
    mode: 'find-files',
    mods: [],
    action: () => {
      pickers.findFiles = false;
    }
  },
]
