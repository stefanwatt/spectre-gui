const offsetLeft = 6
export function calculatePosition(window: App.FloatingWindow): { top: string; left: string } {
  // Base font size in pixels (from your CSS)
  const fontSize = 22;
  // Approximate character width (monospace fonts are typically ~0.6x the height)
  const charWidth = fontSize * 0.7;
  // Line height (from your CSS padding)
  const lineHeight = fontSize + 6; // 3px padding top + 3px padding bottom

  // Calculate position based on row/col
  //
  const top = `${window.row * lineHeight}px`;
  const left = `${window.col+offsetLeft}ch`;

  return { top, left };
}
