import { NeovimConnector } from './neovim-connector';

export class InputHandler {
  private nvim: NeovimConnector;
  
  constructor(nvim: NeovimConnector) {
    this.nvim = nvim;
  }
  
  async handleInput(input: string): Promise<void> {
    await this.nvim.input(input);
  }
  
  async handleMouse(button: string, action: string, modifier: string, grid: number, row: number, col: number): Promise<void> {
    await this.nvim.inputMouse(button, action, modifier, grid, row, col);
  }
  
  async handleCommand(command: string): Promise<void> {
    await this.nvim.command(command);
  }
} 