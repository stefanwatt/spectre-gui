import { pickers } from 'src/routes/state.svelte';
import {
  cursorToNextMatch,
  cursorToPrevMatch,
} from '$lib/picker/results/results.service.svelte';
import { SendKey } from '$lib/wailsjs/go/main/App';

function sendKey(e: KeyboardEvent) {
  e.preventDefault()
  SendKey(e.key, e.ctrlKey, e.altKey, e.shiftKey, 'live-grep');
}

export const keymaps: App.Keymap[] = [
  {
    key: 'Escape',
    mode: 'live-grep',
    mods: [],
    action: () => {
      pickers.liveGrep = false;
    }
  },
  {
    key: 'ArrowDown',
    mode: 'live-grep',
    mods: ['c'],
    action: cursorToNextMatch
  },
  {
    key: 'ArrowUp',
    mode: 'live-grep',
    mods: ['c'],
    action: cursorToPrevMatch
  },
  {
    key: 'ArrowLeft',
    mode: 'live-grep',
    mods: ['c'],
    action: sendKey
  },
  {
    key: 'ArrowRight',
    mode: 'live-grep',
    mods: ['c'],
    action: sendKey
  }
]
