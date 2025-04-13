import { pickers } from '$lib/state.svelte';
import {
  cursorToNextMatch,
  cursorToPrevMatch,
  state as resultsState
} from '$lib/picker/results/results.service.svelte';
import { SendKey,OpenFile } from '$lib/wailsjs/go/main/App';

function sendKey(e: KeyboardEvent) {
  e.preventDefault()
  SendKey(e.key, e.ctrlKey, e.altKey, e.shiftKey, 'live-grep');
}

export const keymaps: App.Keymap[] = [
  {
    key: 'Enter',
    mode: 'live-grep',
    mods: [],
    action: async () => {
      const runtime = await import('$lib/wailsjs/runtime/runtime');
      const selectedMatch = resultsState.selectedMatch
      if (!selectedMatch) return
      if (runtime) {
        OpenFile(selectedMatch.AbsolutePath, selectedMatch.Row, selectedMatch.Col);
      }
    }
  },
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
    mods: [],
    action: cursorToNextMatch
  },
  {
    key: 'ArrowUp',
    mode: 'live-grep',
    mods: [],
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
