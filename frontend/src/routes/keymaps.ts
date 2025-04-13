import { registerKeymap } from 'src/routes/keymap-service';
import { pickers } from 'src/routes/state.svelte';
import {
  cursorToNextMatch,
  cursorToPrevMatch,
} from '$lib/picker/results/results.service.svelte';

function setupKeymaps() {
  registerKeymap({
    key: 'Escape',
    mode: 'live-grep',
    mods: [],
    action: () => {
      pickers.liveGrep = false;
    }
  });

  registerKeymap({
    key: 'ArrowDown',
    mode: 'live-grep',
    mods: ['c'],
    action: cursorToNextMatch
  });
  registerKeymap({
    key: 'ArrowUp',
    mode: 'live-grep',
    mods: ['c'],
    action: cursorToPrevMatch
  });

  registerKeymap({
    key: 'ArrowUp',
    mode: 'live-grep',
    mods: ['c'],
    action: cursorToPrevMatch
  });
}
