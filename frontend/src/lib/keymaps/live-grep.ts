import { nestedState } from '$lib/state.svelte';
import {
  cursorToNextMatch,
  cursorToPrevMatch,
  state
} from '$lib/picker/live-grep-results/results.service.svelte';
import { SendKey, CreateQuickfixList } from '$lib/wailsjs/go/main/App';
import { OpenFile } from '$lib/wailsjs/go/picker/Picker'
import { GetLiveGrepOpts, GetNextPage, GetPrevPage } from '$lib/wailsjs/go/picker/LiveGrepPicker'

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
      const selectedMatch = state.selectedMatch
      if (!selectedMatch) return
      if (runtime) {
        OpenFile(selectedMatch.AbsolutePath, selectedMatch.Row, selectedMatch.Col);
        nestedState.activePicker = undefined
      }
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
