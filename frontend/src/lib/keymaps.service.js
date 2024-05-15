import { get } from "svelte/store";
import { Replace, ReplaceAll, Undo, GetNextPage, GetPrevPage } from "$lib/wailsjs/go/main/App"
import { cursor_to_next_match, cursor_to_prev_match } from "$lib/results/results.service";
import {
  selected_match,
  search_term,
  replace_term,
  dir,
  include,
  exclude,
  case_sensitive,
  regex,
  match_whole_word,
  preserve_case,
  results,
  page_index,
} from "./store";

/** @type {App.Modifier[]}*/
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
      case "ArrowLeft":
        if (is_mod("c")) {
          event.preventDefault()
          GetPrevPage().then(/**@param {App.SearchResult}new_results*/new_results => {
            results.set(new_results.GroupedMatches)
            page_index.set(new_results.PageIndex)
          })
        }
        break

      case "ArrowRight":
        if (is_mod("c")) {
          event.preventDefault()
          GetNextPage().then(/**@param {App.SearchResult}new_results*/new_results => {
            results.set(new_results.GroupedMatches)
            page_index.set(new_results.PageIndex)
          })
        }
        break
      case "ArrowDown":
        event.preventDefault()
        cursor_to_next_match()
        break
      case "ArrowUp":
        event.preventDefault()
        cursor_to_prev_match()
        break

      case "Enter":
        if (is_mod("s")) {
          ReplaceAll(
            get(search_term),
            get(replace_term),
            get(dir),
            get(exclude),
            get(include),
            get(case_sensitive),
            get(regex),
            get(match_whole_word),
            get(preserve_case)
          )
        }
        if (mods.length != 0) return
        Replace(get(selected_match), get(search_term), get(replace_term), get(preserve_case));
        break
      case "i":
        if (!is_mod("a")) return
        case_sensitive.update(old => !old)
        break
      case "w":
        if (!is_mod("a")) return
        // @ts-ignore
        match_whole_word.update(old => !old)
        break
      case "r":
        if (!is_mod("a")) return
        // @ts-ignore
        regex.update(old => !old)
        break
      case "z":
        if (!is_mod("c")) return
        Undo()
        break;
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

/**@param {App.Modifier} mod*/
function is_mod(mod) {
  return mods.includes(mod)
}
