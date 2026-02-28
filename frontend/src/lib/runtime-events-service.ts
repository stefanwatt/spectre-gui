import { OpenFile } from '$lib/wailsjs/go/picker/Picker';
import * as runtime from '$lib/wailsjs/runtime/runtime';
import { windowContentRowMap, layout, cursor, pickers, cmdline, nestedState } from "$lib/state.svelte"
import {
  state as liveGrepState
} from '$lib/picker/live-grep-results/results.service.svelte';
import { handleKeypress } from '$lib/keymaps/keymap-service';

export async function init() {
}
export function startListening() {
  window.addEventListener('keydown', handleKeypress);
  window.addEventListener('resize', function () {
    runtime.EventsEmit('resize');
  });

  runtime.EventsOn('highlight-css', (css: string) => {
    let el = document.getElementById('nvim-hl-style');
    if (!el) {
      el = document.createElement('style');
      el.id = 'nvim-hl-style';
      document.head.appendChild(el);
    }
    el.textContent = css;
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
      liveGrepState.results = updatedResults.GroupedMatches;
      liveGrepState.pageIndex = updatedPageIndex;
    }
  );

  runtime.EventsOn(
    'live-grep-next-page',
    (updatedResults: App.SearchResult, updatedPageIndex: number) => {
      liveGrepState.results = updatedResults.GroupedMatches;
      liveGrepState.pageIndex = updatedPageIndex;
    }
  );

  runtime.EventsOn('live-grep-open-selected-match', () => {
    const selectedMatch = liveGrepState.selectedMatch;
    if (!selectedMatch) return;
    OpenFile(selectedMatch.AbsolutePath, selectedMatch.Row, selectedMatch.Col);
  });

  runtime.EventsOn('BufEnter', (filepath: string) => {
    const activeWindow = layout.windows.find(win => win.id === layout.activeWindowId)
    if (!activeWindow) return
    activeWindow.filepath = filepath
  });

  pickers.forEach(picker => {
    runtime.EventsOn(picker.showEvent, () => {
      nestedState.activePicker = picker
    });
  })

  runtime.EventsOn('cmdline_pos', (data) => {
    cmdline.pos = data.pos;
  });

  runtime.EventsOn('cmdline_hide', (data) => {
    cmdline.visible = false;
  });

  runtime.EventsOn('layout-updated', (updatedLayout: App.NvimLayout) => {
    console.log('layout updates', updatedLayout)
    layout.cols = updatedLayout.cols;
    layout.rows = updatedLayout.rows;
    layout.activeWindowId = updatedLayout?.activeWindowId;
    layout.windows = updatedLayout.windows;
  });

  runtime.EventsOn('content-updated', (winId: number, updatedContent: App.NvimRow[]) => {
    console.log("content-updated for winId=" + winId, updatedContent)
    windowContentRowMap[winId] = updatedContent;
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
    const activeWindow = layout.windows.find((win) => win.id == layout.activeWindowId)
    if (!activeWindow) return
    activeWindow.mode = new_mode
  });

  runtime.EventsOn('floating_windows', (windows: App.FloatingWindow[]) => {
    nestedState.floatingWindows = windows;
  });

  runtime.EventsOn('preview-window', (updatedPreviewWindow: App.FloatingWindow) => {
    if (!nestedState.activePicker) return
    nestedState.previewWindow = updatedPreviewWindow;
  })

  runtime.EventsOn('preview-window-closed', (winId: number) => {
    if (nestedState.previewWindow?.id !== winId) return
    nestedState.previewWindow = undefined;
  })

  runtime.EventsOn('floating_window_closed', (winId: number) => {
    nestedState.floatingWindows = nestedState.floatingWindows.filter((win) => win.id !== winId);
  });

  runtime.EventsOn('hide-window', (winId: number) => {
    delete windowContentRowMap[winId];
  });
}
