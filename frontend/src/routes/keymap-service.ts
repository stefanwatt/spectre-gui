import { SendKey } from '$lib/wailsjs/go/main/App';
import { getKeymapMode } from "./state.svelte"

let activeKeymaps = new Map<string, App.Keymap>()

export function registerKeymap(keymap: App.Keymap) {
  const keymapString = keymapToString(keymap.key, keymap.mods)
  if (activeKeymaps.has(keymapString)) {
    const handler = activeKeymaps.get(keymapString)!.action
    window.removeEventListener("keydown", handler)
    activeKeymaps.delete(keymapString)
  }
  activeKeymaps.set(keymapString, keymap)
}

export function handleKeypress(event: KeyboardEvent) {
  let earlyReturn = false
  const keymapMode = getKeymapMode()
  activeKeymaps.forEach((keymap) => {
    if (keymap.mode !== keymapMode || event.key !== keymap.key || modsPressed(keymap.mods, event)) { return }
    keymap.action(event)
    earlyReturn = true
  })
  if (earlyReturn) return
  event.preventDefault();
  SendKey(event.key, event.ctrlKey, event.altKey, event.shiftKey, keymapMode);
}

function keymapToString(key: string, mods: App.Modifier[]) {
  if (!mods?.length) return key
  return mods.join("-") + "-" + key
}


function modsPressed(mods: App.Modifier[], event: KeyboardEvent): boolean {
  return mods.every(mod => {
    switch (mod) {
      case 'c':
        return event.ctrlKey
      case 's':
        return event.shiftKey
      case 'a':
        return event.altKey
      default:
        break;
    }
  })
}
