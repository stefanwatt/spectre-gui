import LiveGrep from '$lib/picker/LiveGrep.svelte';
import FindFiles from '$lib/picker/FindFiles.svelte';
import FindReferences from '$lib/picker/FindReferences.svelte';
import FindBufferSymbols from '$lib/picker/FindBufferSymbols.svelte';
import { registerKeymap } from './keymaps/keymap-service';

export let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
export let layout = $state<App.NvimLayout>({
  cols: "1fr",
  rows: "1fr",
  activeWindowId: 0,
  windows: []
});
export let nvimWindows = $state<App.NvimWindowMap>({});


export let pickers = $state([
  { showEvent: 'show_live_grep', component: LiveGrep, keymapMode: 'live-grep' },
  { showEvent: 'show-find-files', component: FindFiles, keymapMode: 'find-files' },
  { showEvent: 'show-find-references', component: FindReferences, keymapMode: 'find-references' },
  { showEvent: 'show-find-buffer-symbols', component: FindBufferSymbols, keymapMode: 'find-buffer-symbols' },
] as const satisfies App.Picker[])
pickers.forEach(p => {
  registerKeymap({
    key: 'Escape',
    mode: p.keymapMode,
    mods: [],
    action: () => {
      nestedState.activePicker = undefined
    }
  })
})

export type PickerKeymapMode = typeof pickers[number]['keymapMode'];

export let cmdline = $state<App.CmdLine>({
  visible: false
});
export function getKeymapMode(): App.KeymapMode {
  return keymapMode
}

interface NestedState {
  floatingWindows: App.FloatingWindow[];
  mode: App.VimMode;
  pickerResults: App.PickerResult[]
  activePicker?: App.Picker
}
export let nestedState: NestedState = $state({
  floatingWindows: [],
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
