import { nestedState } from '$lib/state.svelte';
import {
  cursorToNextMatch,
  cursorToPrevMatch,
  state
} from '$lib/picker/live-grep-results/results.service.svelte';
import { CreateQuickfixList } from '@bindings/nvim-gui/app';
import { OpenFile } from '@bindings/nvim-gui/features/picker/picker';
import { GetLiveGrepOpts, GetNextPage, GetPrevPage } from '@bindings/nvim-gui/features/picker/livegreppicker';
import { Call } from '@wailsio/runtime';

function sendKey(e: KeyboardEvent) {
  e.preventDefault()
  Call.ByName('main.App.SendKey', e.key, e.ctrlKey, e.altKey, e.shiftKey, 'live-grep').catch((err) =>
    console.error('SendKey failed', err)
  );
}

export const keymaps: App.Keymap[] = [
  {
    key: 'Enter',
    mode: 'live-grep',
    mods: [],
    action: async () => {
      const selectedMatch = state.selectedMatch
      if (!selectedMatch) return
      OpenFile(selectedMatch.AbsolutePath, selectedMatch.Row, selectedMatch.Col);
      nestedState.activePicker = undefined
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
    action: async () => {
      const liveGrepState = await GetPrevPage()
      state.results = liveGrepState.results
      state.pageIndex = liveGrepState.pageIndex
    }
  },
  {
    key: 'ArrowRight',
    mode: 'live-grep',
    mods: ['c'],
    action: async () => {
      const liveGrepState = await GetNextPage()
      state.results = liveGrepState.results
      state.pageIndex = liveGrepState.pageIndex
    }
  },

  {
    key: 'q',
    mode: 'live-grep',
    mods: ['c'],
    action: async () => {
      const entries = state.results
        .flatMap(result => result.Matches)
        .map(match => {
          return {
            filepath: match.AbsolutePath,
            row: match.Row,
            col: match.Col,
            text: match.MatchedLine
          }
        })
      await CreateQuickfixList(entries)
      nestedState.activePicker = undefined
    }
  },
]
