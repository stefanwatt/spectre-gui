export let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
export let layout = $state<App.NvimLayout>({
  cols: "1fr",
  rows: "1fr",
  activeWindowId: 0,
  windows: []
});
export let nvimWindows = $state<App.NvimWindowMap>({});
export let pickers: App.Pickers = $state({
  liveGrep: false,
  findFiles: false,
  findReferences: false,
});
export let cmdline = $state<App.CmdLine>({
  visible: false
});
let keymapMode: App.KeymapMode = $derived(
  (() => {
    if (cmdline.visible) return 'cmdline';
    if (pickers.liveGrep) return 'live-grep';
    if (pickers.findFiles) return 'find-files';
    if (pickers.findReferences) return 'find-references';
    return 'normal';
  })()
);
export function getKeymapMode(): App.KeymapMode {
  return keymapMode
}

interface NestedState {
  floatingWindows: App.FloatingWindow[];
  mode: App.VimMode;
  pickerResults: App.PickerResult[]
}
export let nestedState: NestedState = $state({
  floatingWindows: [],
  mode: 'normal',
  pickerResults: []
})
