/** @type {App.Modifier[]}*/
let mods = []

/** @type {Map<string, (event:KeyboardEvent)=>void>}*/
let active_keymaps = new Map()

/**@param {App.Keymap[]} keymaps*/
export function setup_keymaps(keymaps) {
  setup_mods()
  keymaps.forEach(keymap => {
    const keymap_string = keymap_to_string(keymap.key, keymap.mods)
    if (active_keymaps.has(keymap_string)) {
      const handler = active_keymaps.get(keymap.key)
      // @ts-ignore 
      window.removeEventListener("keydown", handler)
      active_keymaps.delete(keymap_string)
    }
    window.addEventListener("keydown", (event) => {
      if (event.key !== keymap.key || !keymap.mods.every(is_mod)) { return }
      keymap.action(event)
    })
    active_keymaps.set(keymap_string, keymap.action)
  })
}

/**
 * @param {string} key 
 * @param {App.Modifier[]} mods 
 */
function keymap_to_string(key, mods) {
  if (!mods?.length) return key
  return mods.join("-") + "-" + key
}

function setup_mods() {
  window.addEventListener("keydown", (event) => {
    switch (event.key) {
      case "Control":
        mods.push("c")
        break
      case "Shift":
        mods.push("s")
        break
      case "Alt":
        mods.push("a")
        break
    }
  });

  window.addEventListener("keyup", (event) => {
    switch (event.key) {
      case "Control":
        mods = mods.filter(x => x !== "c")
      case "Shift":
        mods = mods.filter(x => x !== "s")
      case "Alt":
        mods = mods.filter(x => x !== "a")
    }
  });


}
/**@param {App.Modifier} mod*/
function is_mod(mod) {
  return mods.includes(mod)
}
