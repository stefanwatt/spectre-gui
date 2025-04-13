import { OpenFile } from '$lib/wailsjs/go/main/App';
import * as runtime from '$lib/wailsjs/runtime/runtime';
import { nvimWindows, layout, cursor, pickers, cmdline } from "./state.svelte"
import {
  state as resultsState
} from '$lib/picker/results/results.service.svelte';
import { sendKey } from './keymap-service';

export async function init() {
  runtime.EventsEmit('get-highlights');
}
export function startListening() {

  window.addEventListener('keydown', sendKey);

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
      console.log('prev page', updatedResults, updatedPageIndex);
      resultsState.results = updatedResults.GroupedMatches;
      resultsState.pageIndex = updatedPageIndex;
    }
  );

  runtime.EventsOn(
    'live-grep-next-page',
    (updatedResults: App.SearchResult, updatedPageIndex: number) => {
      console.log('next page', updatedResults, updatedPageIndex);
      resultsState.results = updatedResults.GroupedMatches;
      resultsState.pageIndex = updatedPageIndex;
    }
  );

  runtime.EventsOn('live-grep-open-selected-match', () => {
    const selectedMatch = resultsState.selectedMatch;
    if (!selectedMatch) return;
    OpenFile(selectedMatch.AbsolutePath, selectedMatch.Row, selectedMatch.Col);
  });

  runtime.EventsOn('show_live_grep', () => {
    console.log('show live grep');
    pickers.liveGrep = true;
  });

  runtime.EventsOn('hide-live-rep', () => {
    console.log('hide-live-grep');
    pickers.liveGrep = false;
  });

  runtime.EventsOn('cmdline_pos', (data) => {
    cmdline.pos = data.pos;
  });

  runtime.EventsOn('cmdline_hide', (data) => {
    cmdline.visible = false;
  });

  runtime.EventsOn('layout-updated', (updatedLayout: App.NvimLayout) => {
    layout.cols = updatedLayout.cols;
    layout.rows = updatedLayout.cols;
    layout.activeWindowId = updatedLayout?.activeWindowId;
    layout.windows = updatedLayout.windows;
  });

  runtime.EventsOn('content-updated', (winId: number, updatedContent: App.NvimRow[]) => {
    console.log(`content-updated for winId=${winId}`, updatedContent);
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
    //TODO: mode should be on windows not free floating
    // mode = new_mode;
  });

  runtime.EventsOn('floating_windows', (windows: App.FloatingWindow[]) => {
    //TODO:
    // console.log('floating windows', windows);
    // floatingWindows = windows;
  });

  runtime.EventsOn('floating_window_closed', (winId: number) => {
    //TODO:
    // console.log(`floating_window_closed id: ${winId}`);
    // floatingWindows = floatingWindows.filter((win) => win.id !== winId);
  });

  runtime.EventsOn('hide-window', (winId: number) => {
    delete nvimWindows[winId];
  });
}
