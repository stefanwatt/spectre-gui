import MatchWholeWord from "./icons/MatchWholeWord.svelte"
import CaseSensitive from "./icons/CaseSensitive.svelte"
import Regex from "./icons/Regex.svelte"
import { get } from "svelte/store";
import { Replace, ReplaceAll } from "../wailsjs/go/main/App"
import { get_next_match, get_prev_match } from "./results.service";
import { selected_match, results, search_term, replace_term, dir, include, exclude, search_flags, preserve_case } from "./store";
import { CASE_SENSITIVE, MATCH_WHOLE_WORD, REGEX } from "./consts";

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
          ReplaceAll(get(search_term), get(replace_term), get(dir), get(include), get(exclude), get(search_flags))
        }
        if (mods.length != 0) return
        Replace(get(selected_match), get(search_term), get(replace_term));
        break
      case "i":
        if (is_mod("a")) {
          search_flags.update(flags => {
            if (flags.find(flag => flag.text === CASE_SENSITIVE)) {
              return flags.filter(x => x.text !== CASE_SENSITIVE)
            }
            const updated = Array.from(new Set([...flags, { text: CASE_SENSITIVE, icon: CaseSensitive }]))
            console.log("updated flags ", updated)
            return updated
          })
        }
        break
      case "w":
        if (is_mod("a")) {
          search_flags.update(flags => {
            if (flags.find(flag => flag.text === MATCH_WHOLE_WORD)) {
              return flags.filter(x => x.text !== MATCH_WHOLE_WORD)
            }
            const updated = Array.from(new Set([...flags, { text: MATCH_WHOLE_WORD, icon: MatchWholeWord }]))
            console.log("updated flags ", updated)
            return updated
          })
        }
        break
      case "r":
        if (is_mod("a")) {
          search_flags.update(flags => {
            if (flags.find(flag => flag.text === REGEX)) {
              return flags.filter(x => x.text !== REGEX)
            }
            const updated = Array.from(new Set([...flags, { text: REGEX, icon: Regex }]))
            console.log("updated flags ", updated)
            return updated
          })
        }
        break

      case "p":
        if (is_mod("a")) {
          preserve_case.update(x => !x)
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
