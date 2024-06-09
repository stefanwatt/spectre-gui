/**@type{HTMLElement[]}*/
const highlighted_elems = []

const CSS_CLASSES = ['bg-surface1', 'nvim-gui-highlight']
export function clear_highlights() {
  while (highlighted_elems.length) {
    const elem = highlighted_elems.pop()
    CSS_CLASSES.forEach(css_class => {
      elem?.classList?.remove(css_class)
    })
  }
}

/**
 * @param {number} _start_row 
 * @param {number} _end_row 
 */
export function highlight_lines(_start_row, _end_row) {
  let start_row = _start_row
  let end_row = _end_row
  if (_end_row < _start_row) {
    start_row = _end_row
    end_row = _start_row
  }
  clear_highlights()
  for (let row = start_row; row <= end_row; row++) {
    highlight_line(row)
  }
}
/**
 * @param {number} row 
 */
function highlight_line(row) {
  const buf_line_elem = document.getElementById(`buf-line-${row}`)
  if (!buf_line_elem) return;
  CSS_CLASSES.forEach(css_class => {
    buf_line_elem.classList.add(css_class)
    highlighted_elems.push(buf_line_elem)
  })
}

/**
 * @param {App.NvimRange} selection_range 
 * @param {App.BufLine[]} buf_lines
 */
export function highlight_range(selection_range, buf_lines) {
  if (!buf_lines || !selection_range) return;
  let start_row = selection_range.start_row;
  let end_row = selection_range.end_row;
  let start_col = selection_range.start_col;
  let end_col = selection_range.end_col;

  if (selection_range.end_row < selection_range.start_row) {
    start_row = selection_range.end_row
    end_row = selection_range.start_row
    start_col = selection_range.end_col
    end_col = selection_range.start_col
  }

  const cursor = { row: selection_range.end_row, col: selection_range.end_col }
  clear_highlights()
  if (start_row === end_row) {
    for (let col = start_col; col <= end_col; col++) {
      highlight_cell(start_row, col, cursor)
    }
    return;
  }

  const start_buf_line = buf_lines.find(b => b.row === start_row)

  if (!start_buf_line) throw new Error('couldnt fint start_buf_line to highlight')
  const line_end_col = start_buf_line.tokens.slice(-1)[0].end_col;
  for (let col = start_col; col <= line_end_col; col++) {
    highlight_cell(start_row, col, cursor)
  }

  if (end_row - start_row > 1) { //at least one fully highlighted line inbetween
    let row = start_row + 1
    while (row < end_row) {
      highlight_line(row++)
    }
  }

  for (let col = 0; col <= end_col; col++) {
    highlight_cell(end_row, col, cursor)
  }
}

/**
 * @param {number}row
 * @param {number}col
 * @param {App.NvimPosition}cursor
 */
function highlight_cell(row, col, cursor) {
  if (row === cursor.row && col === cursor.col) return;
  const cell_elem = document.getElementById(`cell-${row}-${col}`)
  if (!cell_elem) return;
  CSS_CLASSES.forEach(css_class => {
    cell_elem.classList.add(css_class)
    highlighted_elems.push(cell_elem)
  })
}

