/**@type {NodeJS.Timeout} */
let timer;

/**
 * @param {string} value
 * @returns {Promise<string>}
 */
export function debounce(value) {
  return new Promise((resolve) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      resolve(value);
    }, 400);
  });
}
