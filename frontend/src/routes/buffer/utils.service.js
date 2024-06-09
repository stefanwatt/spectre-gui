/**@param {number} row*/
export function scroll_into_view(row) {
  if (!row) return;
  setTimeout(() => {
    const line_elem = document.getElementById(`buf-line-${row}`);
    if (!line_elem) {
      console.log('couldnt find ' + `.buf-line-${row}`);
      return;
    } else {
      //@ts-ignore
      line_elem.scrollIntoView({ block: 'start' });
    }
  });
	}
