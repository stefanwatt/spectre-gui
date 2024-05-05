import { get } from "svelte/store";
import { get_next_match, get_prev_match } from "./results.service";
import { selected_match, results } from "./store";

/** @type {Modifier[]}*/
let mods = []

export function setup_keymaps() {
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
      case "ArrowDown":
        if (is_mod("c")) {
          const next_match = get_next_match(get(selected_match), get(results))
          if (!next_match) return
          selected_match.set(next_match)
          console.log("selected match:", next_match)
        }
        break
      case "ArrowUp":
        if (is_mod("c")) {
          const next_match = get_prev_match(get(selected_match), get(results))
          if (!next_match) return
          selected_match.set(next_match)
          console.log("selected match:", next_match)
        }
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

/**@param {Modifier} mod*/
function is_mod(mod) {
  return mods.includes(mod)
}
