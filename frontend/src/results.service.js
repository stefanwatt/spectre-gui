/**
 * @returns {RipgrepResult[]}
 * @param {RipgrepResultApi} results
 */
export function map_results(results) {
  if (!results) return [];
  const mapped = Object.entries(results).map(([path, matches]) => {
    return {
      path,
      matches,
    };
  });
  return mapped;
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
