let timer: number

export function debounce<T>(value: T): Promise<T> {
  return new Promise((resolve) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      resolve(value);
    }, 400);
  });
}
