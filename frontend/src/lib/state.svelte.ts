import LiveGrep from '$lib/picker/LiveGrep.svelte';
import FindFiles from '$lib/picker/FindFiles.svelte';
import FindReferences from '$lib/picker/FindReferences.svelte';
import FindHelp from '$lib/picker/FindHelp.svelte';
import FindBufferSymbols from '$lib/picker/FindBufferSymbols.svelte';
import { registerKeymap } from './keymaps/keymap-service';
import { ClosePreview } from '$lib/wailsjs/go/picker/Picker';
import { SvelteMap } from 'svelte/reactivity';

export let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
export let layout = $state<App.NvimLayout>({
  cols: "1fr",
  rows: "1fr",
  activeWindowId: 0,
  windows: []
});
export let windowContentRowMap = new SvelteMap<number, App.NvimRow[]>();


export let pickers = $state([
  { showEvent: 'show_live_grep', component: LiveGrep, keymapMode: 'live-grep' },
  { showEvent: 'show-find-files', component: FindFiles, keymapMode: 'find-files' },
  { showEvent: 'show-find-references', component: FindReferences, keymapMode: 'find-references' },
  { showEvent: 'show-find-buffer-symbols', component: FindBufferSymbols, keymapMode: 'find-buffer-symbols' },
  { showEvent: 'show-find-help', component: FindHelp, keymapMode: 'find-help' },
] as const satisfies App.Picker[])
pickers.forEach(p => {
  registerKeymap({
    key: 'Escape',
    mode: p.keymapMode,
    mods: [],
    action: () => {
      nestedState.activePicker = undefined
      if (nestedState.previewWindow) {
        ClosePreview(nestedState.previewWindow.id)
        nestedState.previewWindow = undefined
      }
    }
  })
})

export type PickerKeymapMode = typeof pickers[number]['keymapMode'];

export let cmdline = $state<App.CmdLine>({
  visible: false
});

export let completion = $state<App.CompletionState>({
  visible: false,
  items: [],
  selectedIndex: -1,
  col: 0,
});

let _completionDocumentation = $state<App.CompletionDocumentation | null>(null);
export function getCompletionDocumentation() { return _completionDocumentation; }
export function setCompletionDocumentation(v: App.CompletionDocumentation | null) { _completionDocumentation = v; }
export function getKeymapMode(): App.KeymapMode {
  return keymapMode
}

let _floatingWindows: App.FloatingWindow[] = $state.raw([]);
export function getFloatingWindows() { return _floatingWindows; }
export function setFloatingWindows(v: App.FloatingWindow[]) { _floatingWindows = v; }

interface NestedState {
  mode: App.VimMode;
  pickerResults: App.PickerResult[]
  activePicker?: App.Picker
  previewWindow?: App.FloatingWindow
}
export let nestedState: NestedState = $state({
  mode: 'normal',
  pickerResults: []
})

let keymapMode: App.KeymapMode = $derived(
  ((): App.KeymapMode => {
    if (cmdline.visible) return 'cmdline';
    if (!nestedState.activePicker) return 'normal';
    return nestedState.activePicker.keymapMode as App.KeymapMode
  })()
);
