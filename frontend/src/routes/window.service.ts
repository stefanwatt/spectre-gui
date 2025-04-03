const offsetLeft = 6
export function calculatePosition(row:number,col:number): { top: string; left: string } {
  const fontSize = 22;
  const lineHeight = fontSize + 6; // 3px padding top + 3px padding bottom
  const top = `${row * lineHeight}px`;
  const left = `${col+offsetLeft}ch`;

  return { top, left };
}
