import { writable, type Writable } from 'svelte/store';

// Define the highlight interface
export interface Highlight {
  id: number;
  fg: string;
  bg: string;
  bold?: boolean;
  italic?: boolean;
  underline?: boolean;
  undercurl?: boolean;
  strikethrough?: boolean;
  reverse?: boolean;
}

// Define the highlights store type
export interface HighlightMap {
  [id: number]: Highlight;
}

// Store for highlight definitions
export const highlights: Writable<HighlightMap> = writable({});

/**
 * Generates CSS for all highlights
 * @param highlightMap - Map of highlight IDs to highlight definitions
 * @returns CSS string with all highlight classes
 */
export function generateHighlightCSS(highlightMap: HighlightMap): string {
  let css = '';
  
  Object.entries(highlightMap).forEach(([id, highlight]) => {
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

/**
 * Updates the style element with new CSS
 * @param css - CSS string to update the style element with
 */
export function updateHighlightStyles(css: string): void {
  let styleEl = document.getElementById('neovim-highlights');
  
  if (!styleEl) {
    styleEl = document.createElement('style');
    styleEl.id = 'neovim-highlights';
    document.head.appendChild(styleEl);
  }
  
  styleEl.textContent = css;
} 