import { OpenFile } from '$lib/wailsjs/go/main/App';
import * as runtime from '$lib/wailsjs/runtime/runtime';
import { nvimWindows, layout, cursor, pickers, cmdline, additionalState } from "$lib/state.svelte"
import {
  state as resultsState
} from '$lib/picker/results/results.service.svelte';
import { handleKeypress } from '$lib/keymaps/keymap-service';

export async function init() {
  runtime.EventsEmit('get-highlights');
}
export function startListening() {
  window.addEventListener('keydown', handleKeypress);
  window.addEventListener('resize', function () {
    runtime.EventsEmit('resize');
  });
  runtime.EventsOn('cmdline_show', (data) => {
    cmdline.visible = true;
    cmdline.content = data.content;
    cmdline.pos = data.pos;
    cmdline.firstc = data.firstc;
    cmdline.prompt = data.prompt;
    cmdline.indent = data.indent;
  });

  runtime.EventsOn(
    'live-grep-prev-page',
    (updatedResults: App.SearchResult, updatedPageIndex: number) => {
      resultsState.results = updatedResults.GroupedMatches;
      resultsState.pageIndex = updatedPageIndex;
    }
  );

  runtime.EventsOn(
    'live-grep-next-page',
    (updatedResults: App.SearchResult, updatedPageIndex: number) => {
      resultsState.results = updatedResults.GroupedMatches;
      resultsState.pageIndex = updatedPageIndex;
    }
  );

  runtime.EventsOn('live-grep-open-selected-match', () => {
    const selectedMatch = resultsState.selectedMatch;
    if (!selectedMatch) return;
    OpenFile(selectedMatch.AbsolutePath, selectedMatch.Row, selectedMatch.Col);
  });

  runtime.EventsOn('show-find-files', () => {
    pickers.findFiles = true;
  });

  runtime.EventsOn('hide-find-files', () => {
    pickers.findFiles = false;
  });


  runtime.EventsOn('show_live_grep', () => {
    pickers.liveGrep = true;
  });

  runtime.EventsOn('hide-live-rep', () => {
    pickers.liveGrep = false;
  });

  runtime.EventsOn('cmdline_pos', (data) => {
    cmdline.pos = data.pos;
  });

  runtime.EventsOn('cmdline_hide', (data) => {
    cmdline.visible = false;
  });

  runtime.EventsOn('layout-updated', (updatedLayout: App.NvimLayout) => {
    console.log('layout updates',updatedLayout)
    layout.cols = updatedLayout.cols;
    layout.rows = updatedLayout.cols;
    layout.activeWindowId = updatedLayout?.activeWindowId;
    layout.windows = updatedLayout.windows;
  });

  runtime.EventsOn('content-updated', (winId: number, updatedContent: App.NvimRow[]) => {
    console.log("content-updated for winId="+winId,updatedContent)
    nvimWindows[winId] = updatedContent;
  });

  runtime.EventsOn(
    'cursor-changed',
    (e: { row: number; col: number; activeWindowId: number }) => {
      cursor.row = e.row;
      cursor.col = e.col;
      if (!layout) return;
      layout.activeWindowId = e.activeWindowId;
    }
  );

  runtime.EventsOn('mode-changed', (new_mode: App.VimMode) => {
    additionalState.mode = new_mode;
  });

  runtime.EventsOn('floating_windows', (windows: App.FloatingWindow[]) => {
    additionalState.floatingWindows = windows;
  });

  runtime.EventsOn('floating_window_closed', (winId: number) => {
    additionalState.floatingWindows = additionalState.floatingWindows.filter((win) => win.id !== winId);
  });

  runtime.EventsOn('hide-window', (winId: number) => {
    delete nvimWindows[winId];
  });
}
