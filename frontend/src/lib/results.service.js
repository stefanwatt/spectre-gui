import { Search } from "$lib/wailsjs/go/main/App";
import { selected_match, results } from "./store";

/**
* @param {string} search_term
* @param {string} dir 
* @param {string} include 
* @param {string} exclude 
* @param {string[]} flags 
* @param {string} replace_term 
* @param {boolean} preserve_case 
 * */
export function search(search_term, dir, include, exclude, flags, replace_term, preserve_case) {
  try {
    Search(
      search_term,
      dir,
      include,
      exclude,
      flags,
      replace_term,
      preserve_case,
    ).then(
        /**@param {App.RipgrepResult[]} res*/(res) => {
        console.log("response:", res)
        selected_match.set(null)
        results.set(res);
        const matches = res[0]?.Matches;
        if (!matches?.length || !matches[0]) {
          selected_match.set(null)
          return;
        }
        const first_match = matches[0];
        console.assert(!!first_match, first_match);
        selected_match.set(first_match)
        setTimeout(() => {
          highlight_all();
        });
      },
    );
  } catch (error) {
    console.error(error)
  }
}

export function highlight_all() {
  if (!window.Prism) return
  window.Prism.highlightAll(false);
}

/** @param {RipgrepMatch}selected_match
 * @param {RipgrepResult[]} results
 * @returns {RipgrepMatch}
 * */
export function get_next_match(selected_match, results) {
  if (!results?.length) return selected_match
  const matches = results.flatMap(result => result.matches)
  if (!matches?.length) return selected_match
  const current_index = matches.indexOf(selected_match)
  if (current_index === -1) return selected_match
  const next_index = current_index + 1
  const last_index = matches.length - 1
  if (next_index > last_index) return matches[0]
  return matches[next_index]
}

/** @param {RipgrepMatch}selected_match
 * @param {RipgrepResult[]} results
 * @returns {RipgrepMatch}
 * */
export function get_prev_match(selected_match, results) {
  if (!results?.length) return selected_match
  const matches = results.flatMap(result => result.matches)
  if (!matches?.length) return selected_match
  const current_index = matches.indexOf(selected_match)
  if (current_index === -1) return selected_match
  const prev_index = current_index - 1
  if (prev_index < 0) return matches[matches.length - 1]
  return matches[prev_index]
}
