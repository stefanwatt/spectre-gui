const nonOffsetWindows= new Set(["treesitter_context" ,"wk","fidget"])
export function calculatePosition(row: number, col: number, filetype: string, anchor: string = 'NW', height: number = 0): { top: string; left: string } {
  const offsetLeft = nonOffsetWindows.has(filetype) ? 0 : 6
  const anchorFontSize = 22;
  const lineHeight = anchorFontSize + 6; // 3px padding top + 3px padding bottom
  // For south anchors (SW/SE), (row, col) is the bottom corner — offset up by window height
  const rowOffset = anchor.startsWith('S') ? (row - height) * lineHeight + 3 : row * lineHeight + 3;
  const top = `${rowOffset}px`;
  const left = `${col + offsetLeft}ch`;

  return { top, left };
}
