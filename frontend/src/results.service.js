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
