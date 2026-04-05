export function createDebounce(ms = 400) {
  let timer: NodeJS.Timeout;
  return function <T>(value: T, overrideMs?: number): Promise<T> {
    return new Promise((resolve) => {
      clearTimeout(timer);
      timer = setTimeout(() => {
        resolve(value);
      }, overrideMs ?? ms);
    });
  };
}

export function isInBounds(element: HTMLElement, parent: HTMLElement): boolean {
  const elementRect = element.getBoundingClientRect();
  if (!parent) parent = window as unknown as HTMLElement
  const parentRect = parent.getBoundingClientRect();

  return (
    elementRect.top >= parentRect.top &&
    elementRect.left >= parentRect.left &&
    elementRect.bottom <= parentRect.bottom &&
    elementRect.right <= parentRect.right
  );
}
