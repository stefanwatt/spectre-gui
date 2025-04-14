const nonOffsetWindows= new Set(["treesitter_context" ,"wk","fidget"])
export function calculatePosition(row: number, col: number, filetype: string): { top: string; left: string } {
  const offsetLeft = nonOffsetWindows.has(filetype) ? 0 : 6
  const anchorFontSize = 22;
  const lineHeight = anchorFontSize + 6; // 3px padding top + 3px padding bottom
  const top = `${row * lineHeight + 3}px`;
  const left = `${col + offsetLeft}ch`;

  return { top, left };
}
