let timer: number

export function debounce<T>(value: T,ms=400): Promise<T> {
  return new Promise((resolve) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      resolve(value);
    }, ms);
  });
}
