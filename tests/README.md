# E2E Tests

End-to-end tests for nvim-gui using Playwright.

## Running the tests

The test script automatically runs inside `nix-shell` to ensure Playwright browsers and dependencies are available:

```bash
bash tests/run-e2e.sh
```

This will:
1. Enter nix-shell if not already in one
2. Kill any existing dev server on port 34115
3. Start `wails dev` with an empty file (`tests/fixtures/empty.go`)
4. Wait for the dev server to accept connections
5. Run Playwright tests
6. Clean up the dev server on exit

## What the tests validate

### Highlighting tests (`highlighting.spec.ts`)
Tests the complete syntax highlighting pipeline:

1. **CSS generation**: Neovim `hl_attr_define` events -> Go highlight parsing -> CSS class generation
2. **CSS injection**: Frontend receives `highlight-css` event and injects into `<style id="nvim-hl-style">` tag
3. **Content rendering**: Grid cells are merged into tokens with highlight classes, rendered as `<span class="cell fg-5 font-bold">` elements
4. **Syntax highlighting**: Different token types (keywords, strings, identifiers) have different `fg-*` classes
5. **CSS consistency**: Every `fg-*`/`bg-*` class on a rendered cell has a matching CSS rule

## Test fixtures

- `tests/fixtures/empty.go` - Empty file used as initial buffer so nvim starts clean
- `tests/fixtures/test.go` - Go file with known syntax highlighting tokens, opened via vim keystrokes during the test

## Architecture notes

- **Problem**: In Wails dev mode, `OnDomReady` fires for the internal webview before external browsers (like Playwright) connect. The initial events (layout, highlights, content) are missed.

| foo | bar | baz | poo | boo |
|-----|-----|-----|-----|-----|
| 1   | 2   | 3   | 4   | 5   |
| 11  | 22  | 33  | 44  | 55  |
| a4  | b4  | c4  | d4  | e4  |

- **Solution**: The frontend emits a `request-state` event on load. The backend responds by re-emitting the current highlights, layout, and content. The test then sends vim keystrokes (`:e tests/fixtures/test.go<CR>`) to open the test file, triggering fresh content-updated events.

