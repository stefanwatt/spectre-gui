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
} from "$lib/store";
import { goto } from "$app/navigation";


/**@type {App.Keymap[]} */
export const keymaps = [
  {
    mods: ["c"],
    key: "ArrowLeft",
    action: (event) => {
      event.preventDefault()
      GetPrevPage().then(/**@param {App.SearchResult}new_results*/new_results => {
        results.set(new_results.GroupedMatches)
        page_index.set(new_results.PageIndex)
      })
    },
  },
  {
    mods: ["c"],
    key: "ArrowRight",
    action: (event) => {
      event.preventDefault()
      GetNextPage().then(/**@param {App.SearchResult}new_results*/new_results => {
        results.set(new_results.GroupedMatches)
        page_index.set(new_results.PageIndex)
      })
    },
  },
  {
    mods: [],
    key: "ArrowDown",
    action: (event) => {
      event.preventDefault()
      cursor_to_next_match()
    },
  },
  {
    mods: [],
    key: "ArrowUp",
    action: (event) => {
      event.preventDefault()
      cursor_to_prev_match()
    },
  },

  {
    mods: ["s"],
    key: "Enter",
    action: (_) => {
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
    },
  },
  {
    mods: [],
    key: "Enter",
    action: (_) => {
      Replace(get(selected_match), get(search_term), get(replace_term), get(preserve_case));
    },
  },
  {
    mods: ["a"],
    key: "i",
    action: (_) => {
      case_sensitive.update(old => !old)
    },
  },
  {
    mods: ["a"],
    key: "w",
    action: (_) => {
      match_whole_word.update(old => !old)
    },
  },
  {
    mods: ["a"],
    key: "r",
    action: (_) => {
      regex.update(old => !old)
    },
  },
  {
    mods: ["a"],
    key: "p",
    action: (_) => {
      preserve_case.update(old => !old)
    },
  },
  {
    mods: ["c"],
    key: "f",
    action: (_) => {
      goto("search")
    }
  },
  {
    mods: ["c"],
    key: "z",
    action: (_) => {
      Undo()
    }
  },
]
