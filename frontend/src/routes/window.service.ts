const nonOffsetWindows= new Set(["treesitter_context" ,"wk","fidget"])
export function calculatePosition(row: number, col: number, filetype: string): { top: string; left: string } {
  const offsetLeft = nonOffsetWindows.has(filetype) ? 0 : 6
  console.log("calculatePosition filetype=" + filetype + " offsetLeft=" + offsetLeft)
  const fontSize = 22;
  const lineHeight = fontSize + 6; // 3px padding top + 3px padding bottom
  const top = `${row * lineHeight}px`;
  const left = `${col + offsetLeft}ch`;

  return { top, left };
}
