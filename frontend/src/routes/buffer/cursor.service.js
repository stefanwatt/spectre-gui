
const LINE_END_CLASSES = ['!border-r-rosewater', 'border-transparent', 'border-r']
const CURSOR_CELL_CLASSES_I = ["border-l-rosewater"]
const CURSOR_CELL_CLASSES = ["bg-rosewater", "text-mantle"]
/**@type{HTMLElement|null}*/
let cursor_cell;
/**@type{HTMLElement|null}*/
let line_end_token;
/**
 * @param {App.NvimPosition} updated_cursor
 * @param {number} line_end_col
 * @param {App.VimMode} mode
 */
export function update_cursor(updated_cursor, line_end_col, mode) {
  console.log("update cursor", updated_cursor)
  const css_classes = mode === "i" ? CURSOR_CELL_CLASSES_I : CURSOR_CELL_CLASSES
  const unused_css_classes = mode !== "i" ? CURSOR_CELL_CLASSES_I : CURSOR_CELL_CLASSES
  if (mode === 'i' && updated_cursor.col >= line_end_col) {
    return update_line_end_cursor(updated_cursor, css_classes, unused_css_classes)
  }

  const updated_cursor_cell = document.getElementById(`cell-${updated_cursor.row}-${updated_cursor.col}`);
  if (updated_cursor_cell && cursor_cell) {
    css_classes.forEach(css_class => cursor_cell?.classList?.remove(css_class))
  }
  if (!updated_cursor_cell) {
    cursor_cell = null;
    return;
  }
  if (!updated_cursor_cell.classList) {
    updated_cursor_cell.className = css_classes.join(" ");
  } else {
    css_classes.forEach(css_class => updated_cursor_cell?.classList?.add(css_class))
    unused_css_classes.forEach(css_class => updated_cursor_cell?.classList?.remove(css_class))
  }
  LINE_END_CLASSES.forEach(css_class => line_end_token?.classList?.remove(css_class))
  line_end_token = null;
  cursor_cell = updated_cursor_cell;
}

/**
 * @param {App.NvimPosition} updated_cursor 
 * @param {string[]}css_classes
 *@param {string[]}unused_css_classes
 * */
function update_line_end_cursor(updated_cursor, css_classes, unused_css_classes) {
  /**@type{HTMLElement}*/
  //@ts-ignore
  const updated_cursor_line = document.getElementById(`buf-line-${updated_cursor.row}`);
  /**@type{HTMLElement}*/
  //@ts-ignore
  const updated_line_end_token = Array.from(updated_cursor_line.querySelectorAll(".nvim-gui-token")).slice(-1)[0]
  if (updated_line_end_token && line_end_token?.classList) {
    LINE_END_CLASSES.forEach(css_class => line_end_token?.classList?.remove(css_class))
  }
  if (!updated_line_end_token) {
    line_end_token = null;
    return;
  }
  if (!updated_line_end_token.classList) {
    updated_cursor_line.className = LINE_END_CLASSES.join(" ")
  } else {
    LINE_END_CLASSES.forEach(css_class => updated_line_end_token?.classList?.add(css_class))
  }
  css_classes.forEach(css_class => cursor_cell?.classList?.remove(css_class))
  unused_css_classes.forEach(css_class => cursor_cell?.classList?.remove(css_class))
  cursor_cell = null;
  line_end_token = updated_line_end_token
}
