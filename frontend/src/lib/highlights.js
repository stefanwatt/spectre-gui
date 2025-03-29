import { writable } from 'svelte/store';

// Store for highlight definitions
export const highlights = writable(new Map());

// Function to generate CSS for all highlights
/**
 * @param {Map<string, {fg: string, bg: string, bold: boolean, italic: boolean, underline: boolean, strikethrough: boolean, undercurl: boolean, reverse: boolean}>} highlightMap
 * @returns {string}
 */
export function generateHighlightCSS(highlightMap) {
  let css = '';
  
  Object.entries(highlightMap).forEach(
    /**
     * @param {[string, {fg: string, bg: string, bold: boolean, italic: boolean, underline: boolean, strikethrough: boolean, undercurl: boolean, reverse: boolean}]} id
     */
    ([id, highlight]) => {
    css += `.hl-${id} {\n`;
    
    if (highlight.fg) {
      css += `  color: ${highlight.fg};\n`;
    }
    
    if (highlight.bg) {
      css += `  background-color: ${highlight.bg};\n`;
    }
    
    if (highlight.bold) {
      css += `  font-weight: bold;\n`;
    }
    
    if (highlight.italic) {
      css += `  font-style: italic;\n`;
    }
    
    if (highlight.underline) {
      css += `  text-decoration: underline;\n`;
    }
    
    if (highlight.strikethrough) {
      css += `  text-decoration: line-through;\n`;
    }
    
    if (highlight.undercurl) {
      css += `  text-decoration: underline wavy;\n`;
    }
    
    if (highlight.reverse) {
      // For reverse, we swap foreground and background
      css += `  color: ${highlight.bg || 'inherit'};\n`;
      css += `  background-color: ${highlight.fg || 'inherit'};\n`;
    }
    
    css += `}\n`;
  });
  
  return css;
}

// Function to update the style element with new CSS
/** @param {string} css*/
export function updateHighlightStyles(css) {
  console.log("updating css: ",css)
  let styleEl = document.getElementById('neovim-highlights');
  
  if (!styleEl) {
    styleEl = document.createElement('style');
    styleEl.id = 'neovim-highlights';
    document.head.appendChild(styleEl);
  }
  
  styleEl.textContent = css;
} 
