import { get } from "svelte/store";
import { Replace, ReplaceAll } from "../wailsjs/go/main/App"
import { get_next_match, get_prev_match } from "./results.service";
import { selected_match, results, search_term, replace_term, dir, include, exclude } from "./store";

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

      case "Enter":
        if (is_mod("s")) {
          ReplaceAll(get(search_term), get(replace_term), get(dir), get(include), get(exclude))
        }
        if (mods.length != 0) return
        Replace(get(selected_match), get(search_term), get(replace_term));

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
