import { LiveGrep } from "$lib/wailsjs/go/picker/LiveGrepPicker";
export interface ResultsState {
  selectedMatch: App.RipgrepMatch | null;
  results: App.RipgrepResult[];
  toast: App.Toast | null;
  totalPages: number;
  pageIndex: number;
  totalResults: number;
  totalFiles: number;
}

export let state: ResultsState = $state({
  selectedMatch: null,
  results: [],
  toast: null,
  totalPages: 0,
  pageIndex: 0,
  totalResults: 0,
  totalFiles: 0
})

export function search(
  searchTerm: string,
  dir: string,
  exclude: string,
  include: string,
  caseSensitive: boolean,
  regex: boolean,
  matchWholeWord: boolean) {
  try {
    console.log("searching for: " + searchTerm)
    LiveGrep(
      searchTerm,
      dir,
      exclude,
      include,
      caseSensitive,
      regex,
      matchWholeWord,
    ).then(
      (res) => {
        console.log("response:", res)
        state.selectedMatch = null
        state.results = res.results;
        if (!res.results?.length) {
          state.totalFiles = 0
          state.totalResults = 0
          state.pageIndex = 0
          state.totalPages = 0
          return;
        }
        const matches = res.results[0]?.Matches;
        if (!matches?.length || !matches[0]) {
          state.selectedMatch = null
          return;
        }
        const first_match = matches[0];
        console.assert(!!first_match, first_match);
        state.selectedMatch = first_match
        state.totalFiles = res.totalFiles
        state.totalResults = res.totalResults
        state.pageIndex = res.pageIndex
        state.totalPages = res.totalPages
      },
    );
  } catch (error) {
    console.error(error)
  }
}

export function getNextMatch(selected_match: App.RipgrepMatch, results: App.RipgrepResult[]): App.RipgrepMatch {
  console.log("get_next_match", selected_match, results)
  if (!results?.length) return selected_match
  const matches = new Map()
  const matchesList = results.flatMap(result => result.Matches)
  for (let i = 0; i < matchesList.length; i++) {
    const match = matchesList[i];
    matches.set(match.Id, { index: i, match })
  }
  if (!matches.size) return selected_match
  if (!matches.has(selected_match.Id)) return selected_match
  const current_index = matches.get(selected_match.Id).index
  const next_index = current_index + 1
  const last_index = matches.size - 1
  if (next_index > last_index) return matchesList[0]
  return matchesList[next_index]
}

export function get_prev_match(selected_match: App.RipgrepMatch, results: App.RipgrepResult[]): App.RipgrepMatch {
  if (!results?.length) return selected_match
  const matches = results.flatMap(result => result.Matches)
  if (!matches?.length) return selected_match
  const current_index = matches.indexOf(selected_match)
  if (current_index === -1) return selected_match
  const prev_index = current_index - 1
  if (prev_index < 0) return matches[matches.length - 1]
  return matches[prev_index]
}


export function cursorToNextMatch() {
  console.log("cursor_to_next_match")
  let current_match = state.selectedMatch
  if (!current_match) {
    current_match = state.results[0]?.Matches[0]
    if (!current_match) return
    state.selectedMatch = current_match
    console.log("selected match:", current_match)
  }
  const next_match = getNextMatch(current_match, state.results)
  if (!next_match) return
  state.selectedMatch = next_match
  console.log("selected match:", next_match)
}

export function cursorToPrevMatch() {
  console.log("cursor_to_prev_match")
  let current_match = state.selectedMatch
  if (!current_match) {
    current_match = state.results[0]?.Matches[0]
    if (!current_match) return
    state.selectedMatch = current_match
    console.log("selected match:", current_match)
  }
  const prev_match = get_prev_match(current_match, state.results)
  if (!prev_match) return
  state.selectedMatch = prev_match
  console.log("selected match:", prev_match)
}

