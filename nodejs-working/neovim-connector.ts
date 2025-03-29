import { spawn } from 'child_process';
import { NeovimClient } from 'neovim';
import { Readable, Writable } from 'stream';

import { TerminalRenderer } from './terminal-renderer.js';

export class NeovimConnector {
  private nvim: NeovimClient;
  private renderer: TerminalRenderer;
  private redrawBatch: any[][] = [];
  private batchTimeout: NodeJS.Timeout | null = null;

  constructor(nvim: NeovimClient, renderer: TerminalRenderer) {
    this.nvim = nvim;
    this.renderer = renderer;

    // Set up event handlers
    this.nvim.on('notification', this.handleNotification);
    this.nvim.on('request', this.handleRequest);
    this.nvim.on('disconnect', this.handleDisconnect);
  }

  static async connect(type: string, path: string, filepath: string): Promise<NeovimClient> {
    console.log(`Connecting to Neovim with type: ${type}, path: ${path}`);

    if (type === 'command') {
      try {
        console.log('Spawning Neovim process...');
        const proc = spawn(path, ['--embed', filepath], {
          stdio: ['pipe', 'pipe', process.stderr]
        });

        // Add event handlers for the process
        proc.on('error', (err) => {
          console.error('Neovim process error:', err);
        });

        proc.on('exit', (code, signal) => {
          console.log(`Neovim process exited with code ${code} and signal ${signal}`);
        });

        if (!proc.stdout || !proc.stdin) {
          throw new Error('Failed to get stdout or stdin from Neovim process');
        }

        console.log('Creating NeovimClient...');
        const nvim = new NeovimClient();

        console.log('Attaching to Neovim process...');
        await nvim.attach({
          reader: proc.stdout as Readable,
          writer: proc.stdin as Writable
        });

        console.log('Successfully attached to Neovim process');
        return nvim;
      } catch (err) {
        console.error('Error connecting to Neovim:', err);
        throw err;
      }
    } else {
      throw new Error(`Connection type ${type} not supported yet`);
    }
  }

  async attach(cols: number, rows: number): Promise<void> {
    console.log(`Attaching UI with dimensions: ${cols}x${rows}`);

    try {
      // Attach to Neovim UI with required extensions
      await this.nvim.uiAttach(cols, rows, {
        rgb: true,
        ext_linegrid: true,
        ext_multigrid: true,
        ext_cmdline: true,
        ext_popupmenu: true,
        ext_tabline: true,
        ext_messages: true
      });

      // Set client info
      await this.nvim.setClientInfo(
        'Terminal-Envim',
        { major: 0, minor: 1, patch: 0 },
        'ui',
        {},
        {}
      );

      console.log('UI attached successfully');
    } catch (err) {
      console.error('Error attaching UI:', err);
      throw err;
    }
  }

  private handleNotification = (method: string, args: any[]): void => {
    if (method === 'redraw') {
      this.redraw(args);
    }
  }

  private handleRequest = (method: string, args: any[], resp: any): void => {
    resp.send(null, 'ok');
  }

  private handleDisconnect = (): void => {
    console.log('Neovim disconnected');
    process.exit(0);
  }

  private redraw = (events: any[][]): void => {
    events.forEach(args => {
      const name = args.shift();
      switch (name) {
        case 'grid_resize':
          args.forEach(arg => this.renderer.gridResize(arg[0], arg[1], arg[2]));
          break;
        case 'grid_line':
          args.forEach(arg => this.renderer.gridLine(arg[0], arg[1], arg[2], arg[3]));
          break;
        case 'flush':
          this.renderer.scheduleRender();
          break;
        case 'grid_clear':
          args.forEach(arg => this.renderer.gridClear(arg[0]));
          break;
        case 'grid_cursor_goto':
          args.forEach(arg => this.renderer.gridCursorGoto(arg[0], arg[1], arg[2]));
          break;
        case 'grid_scroll':
          args.forEach(arg => this.renderer.gridScroll(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5], arg[6]));
          break;
        case 'default_colors_set':
          args.forEach(arg => this.renderer.defaultColorsSet(arg[0], arg[1], arg[2]));
          break;
        case 'hl_attr_define':
          this.renderer.hlAttrDefine(args);
          break;
        case 'mode_change':
          args.forEach(arg => this.renderer.modeChange(arg[0], arg[1]));
          break;
        case 'win_pos':
          args.forEach(arg => this.renderer.winPos(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5]));
          break;
        case 'win_float_pos':
          args.forEach(arg => this.renderer.winFloatPos(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5], arg[6], arg[7]));
          break;
      }
    })
  }

  async input(input: string): Promise<void> {
    await this.nvim.input(input);
  }

  async inputMouse(button: string, action: string, modifier: string, grid: number, row: number, col: number): Promise<void> {
    await this.nvim.inputMouse(button, action, modifier, grid, row, col);
  }

  async command(command: string): Promise<void> {
    await this.nvim.command(command);
  }
}
