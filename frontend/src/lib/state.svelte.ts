export let cursor = $state<App.NvimPosition>({ row: 0, col: 1 });
export let layout = $state<App.NvimLayout>({
  cols: "1fr",
  rows: "1fr",
  activeWindowId: 0,
  windows: []
});
export let nvimWindows = $state<App.NvimWindowMap>({});
export let pickers = $state({ liveGrep: false, findFiles: false });
export let cmdline = $state<App.CmdLine>({
  visible: false
});
let keymapMode: App.KeymapMode = $derived(
  (() => {
    if (cmdline.visible) return 'cmdline';
    if (pickers.liveGrep) return 'live-grep';
    if (pickers.findFiles) return 'find-files';
    return 'normal';
  })()
);
export function getKeymapMode(): App.KeymapMode {
  return keymapMode
}
export let additionalState: { floatingWindows: App.FloatingWindow[], mode: App.VimMode } = $state({
  floatingWindows: [],
  mode: 'normal'
})
