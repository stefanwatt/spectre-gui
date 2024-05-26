import { get } from "svelte/store";
import { Replace, ReplaceAll, GetNextPage, GetPrevPage, OpenMatch } from "$lib/wailsjs/go/main/App"
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
    mods: [],
    key: "Enter",
    action: (_) => {
      const match = get(selected_match)
      if (!match) { return; }
      OpenMatch(match.AbsolutePath, match.Row, match.Col)
    },
  },
  {
    mods: ["s"],
    key: "Enter",
    action: (_) => {
      //TODO open file
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
    mods: ["c"],
    key: "h",
    action: (_) => {
      goto("search-and-replace")
    }
  },
]
