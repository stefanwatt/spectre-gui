let timer: NodeJS.Timeout

export function debounce<T>(value: T, ms = 400): Promise<T> {
  return new Promise((resolve) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      resolve(value);
    }, ms);
  });
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
