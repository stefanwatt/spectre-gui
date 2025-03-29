import blessed from 'blessed';

import { writeFileSync } from "fs"
import { TerminalRenderer } from './terminal-renderer.js';
import { NeovimConnector } from './neovim-connector.js';
import { InputHandler } from './input-handler.js';

let screen, renderer, nvim, connector, inputHandler;

async function main() {
  writeFileSync("envim.log", "")
  try {
    // Create blessed screen
    console.log('Creating screen...');
    screen = blessed.screen({
      smartCSR: true,
      title: 'Terminal Envim'
    });

    // Handle exit
    screen.key(['C-c'], () => {
      console.log('Exiting...');
      return process.exit(0);
    });

    // Create terminal renderer
    console.log('Creating terminal renderer...');
    renderer = new TerminalRenderer(screen);

    // Connect to Neovim
    console.log('Connecting to Neovim...');
    try {
      nvim = await NeovimConnector.connect('command', 'nvim', '/tmp/foo.lua');
      console.log('Connected to Neovim');
    } catch (err) {
      console.error('Failed to connect to Neovim:', err);
      throw err;
    }

    // Initialize the UI
    const { cols, rows } = renderer.getDimensions();
    console.log(`Terminal dimensions: ${cols}x${rows}`);

    // Set up event handlers
    console.log('Setting up event handlers...');
    connector = new NeovimConnector(nvim, renderer);

    // Create input handler and connect it to the renderer
    console.log('Setting up input handler...');
    inputHandler = new InputHandler(connector);
    renderer.setInputHandler(inputHandler);

    // Attach to Neovim UI
    console.log(`Attaching UI with dimensions: ${cols}x${rows}`);
    try {
      await connector.attach(cols, rows);
      console.log('UI attached successfully');
    } catch (err) {
      console.error('Failed to attach UI:', err);
      throw err;
    }

    // Render the screen
    console.log('Rendering screen...');
    screen.render();

    // Keep the process alive
    console.log('Terminal Envim is running. Press Ctrl+C to exit.');

    // Add a simple interval to keep the process alive
    setInterval(() => {
      // Do nothing, just keep the process alive
    }, 1000);

  } catch (err) {
    console.error('Error during initialization:', err);
    process.exit(1);
  }
}

// Add global error handlers
process.on('uncaughtException', (err) => {
  console.error('Uncaught exception:', err);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled rejection at:', promise, 'reason:', reason);
});

main().catch(err => {
  console.error('Error in main function:', err);
  process.exit(1);
});
