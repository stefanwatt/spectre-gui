import blessed from 'blessed';
import { InputHandler } from './input-handler.js';

interface IHighlight {
  foreground?: number;
  background?: number;
  special?: number;
  reverse?: boolean;
  italic?: boolean;
  bold?: boolean;
  underline?: boolean;
  undercurl?: boolean;
  strikethrough?: boolean;
}

interface IGrid {
  id: number;
  width: number;
  height: number;
  cells: {
    text: string;
    hl: string;
  }[][];
  cursor?: {
    row: number;
    col: number;
  };
}

export class TerminalRenderer {
  private screen: blessed.Widgets.Screen;
  private grids: Map<number, IGrid> = new Map();
  private highlights: Map<string, IHighlight> = new Map();
  private defaultFg: number = 0xffffff;
  private defaultBg: number = 0x000000;
  private defaultSp: number = 0xffffff;
  private activeGrid: number = 1;
  private mode: string = 'normal';
  private inputHandler?: InputHandler;
  private pendingRender: boolean = false;

  constructor(screen: blessed.Widgets.Screen) {
    this.screen = screen;

    // Create default grid (1)
    this.grids.set(1, {
      id: 1,
      width: +screen.width,
      height: +screen.height,
      cells: Array(+screen.height).fill(0).map(() =>
        Array(+screen.width).fill(0).map(() => ({ text: ' ', hl: '0' }))
      )
    });

    // Set default highlight
    this.highlights.set('0', {
      foreground: this.defaultFg,
      background: this.defaultBg,
      special: this.defaultSp
    });

    // Handle key events
    this.screen.on('keypress', (ch, key) => {
      if (!key) return;

      if (key.name === 'escape') {
        this.handleInput('<Esc>');
      } else if (key.ctrl) {
        this.handleInput(`<C-${key.name}>`);
      } else if (key.meta) {
        this.handleInput(`<A-${key.name}>`);
      } else if (key.name === 'return') {
        this.handleInput('<CR>');
      } else if (key.name === 'backspace') {
        this.handleInput('<BS>');
      } else if (key.name === 'space') {
        this.handleInput(' ');
      } else if (key.name === 'tab') {
        this.handleInput('<Tab>');
      } else if (key.name === 'up') {
        this.handleInput('<Up>');
      } else if (key.name === 'down') {
        this.handleInput('<Down>');
      } else if (key.name === 'left') {
        this.handleInput('<Left>');
      } else if (key.name === 'right') {
        this.handleInput('<Right>');
      } else if (ch) {
        this.handleInput(ch);
      }
    });
  }

  setInputHandler(inputHandler: InputHandler): void {
    this.inputHandler = inputHandler;
  }

  getDimensions(): { cols: number, rows: number } {
    return {
      cols: +this.screen.width,
      rows: +this.screen.height
    };
  }

  private handleInput(input: string): void {
    if (this.inputHandler) {
      this.inputHandler.handleInput(input).catch(err => {
        console.error('Error handling input:', err);
      });
    }
  }

  gridResize(gridId: number, width: number, height: number): void {
    const grid = this.grids.get(gridId) || {
      id: gridId,
      width: 0,
      height: 0,
      cells: []
    };

    grid.width = width;
    grid.height = height;

    // Resize the grid cells
    grid.cells = Array(height).fill(0).map((_, row) => {
      const existingRow = grid.cells[row] || [];
      return Array(width).fill(0).map((_, col) => {
        return existingRow[col] || { text: ' ', hl: '0' };
      });
    });

    this.grids.set(gridId, grid);
    this.scheduleRender();
  }

  gridLine(gridId: number, row: number, col: number, cells: any[]): void {
    const grid = this.grids.get(gridId);
    if (!grid) return;

    let currentCol = col;
    for (const cell of cells) {
      const text = cell[0];
      const hlId = cell[1] || '0';
      const width = cell[2] || 1;

      if (row < grid.height && currentCol < grid.width) {
        grid.cells[row][currentCol] = { text, hl: hlId };
      }

      currentCol += width;
    }
  }

  gridClear(gridId: number): void {
    const grid = this.grids.get(gridId);
    if (!grid) return;

    for (let row = 0; row < grid.height; row++) {
      for (let col = 0; col < grid.width; col++) {
        grid.cells[row][col] = { text: ' ', hl: '0' };
      }
    }

    this.scheduleRender();
  }

  gridCursorGoto(gridId: number, row: number, col: number): void {
    const grid = this.grids.get(gridId);
    if (!grid) return;

    grid.cursor = { row, col };
    this.activeGrid = gridId;
    // log(`gridCursorGoto: gridId:${gridId}, row:${row}, col:${col}, grid: ${renderGrid(grid)}`)
    this.scheduleRender();
  }

  gridScroll(gridId: number, top: number, bot: number, left: number, right: number, rows: number, cols: number): void {
    // log(`gridScroll id:${gridId}, top:${top}, bot:${bot}, left:${left}, right:${right}, rows:${rows}, cols:${cols} `)
    const grid = this.grids.get(gridId);
    if (!grid) return;

    // Handle vertical scrolling
    if (rows !== 0) {
      const direction = rows > 0 ? 1 : -1;
      const absRows = Math.abs(rows);

      if (direction > 0) {
        // Scroll down
        for (let row = bot - 1; row >= top + absRows; row--) {
          for (let col = left; col < right; col++) {
            if (row < grid.height && col < grid.width && row - absRows >= 0) {
              grid.cells[row][col] = grid.cells[row - absRows][col];
            }
          }
        }
        // Clear the newly exposed area
        for (let row = top; row < top + absRows && row < grid.height; row++) {
          for (let col = left; col < right && col < grid.width; col++) {
            grid.cells[row][col] = { text: ' ', hl: '0' };
          }
        }
      } else {
        // Scroll up
        for (let row = top; row < bot - absRows; row++) {
          for (let col = left; col < right; col++) {
            if (row < grid.height && col < grid.width && row + absRows < grid.height) {
              grid.cells[row][col] = grid.cells[row + absRows][col];
            }
          }
        }
        // Clear the newly exposed area
        for (let row = bot - absRows; row < bot && row < grid.height; row++) {
          for (let col = left; col < right && col < grid.width; col++) {
            grid.cells[row][col] = { text: ' ', hl: '0' };
          }
        }
      }
    }

    this.scheduleRender();
  }

  defaultColorsSet(fg: number, bg: number, sp: number): void {
    this.defaultFg = fg >= 0 ? fg : 0xffffff;
    this.defaultBg = bg >= 0 ? bg : 0x000000;
    this.defaultSp = sp >= 0 ? sp : this.defaultFg;

    this.highlights.set('0', {
      foreground: this.defaultFg,
      background: this.defaultBg,
      special: this.defaultSp
    });

    this.scheduleRender();
  }

  hlAttrDefine(hlAttrs: any[]): void {
    for (const attr of hlAttrs) {
      const id = attr[0];
      const rgbAttrs = attr[1];

      this.highlights.set(id.toString(), {
        foreground: rgbAttrs.foreground,
        background: rgbAttrs.background,
        special: rgbAttrs.special,
        reverse: rgbAttrs.reverse,
        italic: rgbAttrs.italic,
        bold: rgbAttrs.bold,
        underline: rgbAttrs.underline,
        undercurl: rgbAttrs.undercurl,
        strikethrough: rgbAttrs.strikethrough
      });
    }

    this.scheduleRender();
  }

  modeChange(mode: string, modeIdx: number): void {
    this.mode = mode;
    this.scheduleRender();
  }

  winPos(gridId: number, win: any, row: number, col: number, width: number, height: number): void {
    // Handle window positioning
    if (!this.grids.has(gridId)) {
      this.gridResize(gridId, width, height);
    }
  }

  winFloatPos(gridId: number, win: any, anchor: string, anchorGrid: number, anchorRow: number, anchorCol: number, focusable: boolean, zIndex: number): void {
    // Handle floating window positioning
    if (!this.grids.has(gridId)) {
      const width = 10; // Default width
      const height = 5; // Default height
      this.gridResize(gridId, width, height);
    }
  }

  msgShow(messages: any[]): void {
    // Handle message display
    console.log('Message:', messages);
  }

  public scheduleRender(): void {
    if (!this.pendingRender) {
      this.pendingRender = true;
      // Use setImmediate to batch multiple updates in the same event loop tick
      setImmediate(() => {
        this.render();
        this.pendingRender = false;
      });
    }
  }

  render(): void {
    // Use a buffer to build the entire screen content
    const buffer = Array(+this.screen.height).fill(0).map(() =>
      Array(+this.screen.width).fill({ text: ' ', hl: '0' })
    );

    // Render all visible grids into the buffer
    for (const [gridId, grid] of this.grids.entries()) {
      if (gridId !== 2) continue
      // Copy grid cells to buffer
      for (let row = 0; row < grid.height && row < buffer.length; row++) {
        for (let col = 0; col < grid.width && col < buffer[row].length; col++) {
          if (row < grid.cells.length && col < grid.cells[row].length) {
            buffer[row][col] = grid.cells[row][col];
          }
        }
      }
    }

    // Clear the screen once
    this.screen.clearRegion(0, +this.screen.width, 0, +this.screen.height);

    // Render the buffer to the screen
    for (let row = 0; row < buffer.length; row++) {
      let line = '';
      let currentHl = '';

      for (let col = 0; col < buffer[row].length; col++) {
        const cell = buffer[row][col];
        const hlId = cell.hl;

        // If highlight changes, flush the current line segment
        if (hlId !== currentHl) {
          if (line.length > 0) {
            // Render the current line segment with its highlight
            this.renderLineSegment(row, col - line.length, line, currentHl);
            line = '';
          }
          currentHl = hlId;
        }

        // Add character to the current line segment
        line += cell.text || ' ';
      }

      // Render any remaining text in the line
      if (line.length > 0) {
        this.renderLineSegment(row, buffer[row].length - line.length, line, currentHl);
      }
    }

    // Render the cursor
    const activeGrid = this.grids.get(this.activeGrid);
    if (activeGrid && activeGrid.cursor) {
      this.screen.program.cursorPos(activeGrid.cursor.row, activeGrid.cursor.col);
      this.screen.program.showCursor();
    }

    // Refresh the screen
    this.screen.render();
  }

  private renderLineSegment(row: number, col: number, text: string, hlId: string): void {
    const hl = this.highlights.get(hlId) || this.highlights.get('0')!;

    // Convert RGB colors to terminal colors
    const fg = this.rgbToTerminal(hl.foreground || this.defaultFg);
    const bg = this.rgbToTerminal(hl.background || this.defaultBg);

    // Position cursor at the start of the segment
    this.screen.program.cursorPos(row, col);

    // Set colors and attributes
    this.screen.program.bg(bg.toString());
    this.screen.program.fg(fg.toString());

    if (hl.reverse) this.screen.program.reverse();

    // Write the text segment
    this.screen.program.write(text);

    // Reset attributes
    this.screen.program.bg(this.rgbToTerminal(this.defaultBg).toString());
    this.screen.program.fg(this.rgbToTerminal(this.defaultFg).toString());
    if (this.screen.program.resetCursor) this.screen.program.resetCursor();
  }

  private rgbToTerminal(rgb: number): number {
    if (rgb === undefined) return 0;

    // Extract RGB components
    const r = (rgb >> 16) & 0xFF;
    const g = (rgb >> 8) & 0xFF;
    const b = rgb & 0xFF;

    // Convert to 0-5 range for each component
    const tr = Math.round(r * 5 / 255);
    const tg = Math.round(g * 5 / 255);
    const tb = Math.round(b * 5 / 255);

    // Calculate the color index (16 + 36*r + 6*g + b)
    return 16 + 36 * tr + 6 * tg + tb;
  }
}
