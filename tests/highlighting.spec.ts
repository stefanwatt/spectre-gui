import { test, expect, type Page } from '@playwright/test';

// These tests validate the highlighting and rendering pipeline end-to-end:
// 1. Neovim starts with empty.go
// 2. Test sends vim keystrokes to open test.go (:e tests/fixtures/test.go<CR>)
// 3. This triggers fresh content-updated events that the test browser captures
// 4. Neovim hl_attr_define → Go highlight parsing → CSS class generation
// 5. Frontend receives highlight-css event and injects into <style> tag
// 6. Grid cells are merged into tokens with class strings (grid_line events come incrementally)
// 7. Frontend receives content-updated events and renders <span> elements

let page: Page;

test.describe.configure({ mode: 'serial' });

test.beforeAll(async ({ browser }) => {
  page = await browser.newPage();
  await page.goto('/');

  // Wait for the highlight style tag to appear (nvim has started)
  await expect(page.locator('style#nvim-hl-style')).toBeAttached({ timeout: 15000 });

  // Send vim command to open test.go - this triggers content-updated events
  // that the test browser will capture
  await page.keyboard.type(':e tests/fixtures/test.go');
  await page.keyboard.press('Enter');

  // Wait for content to start appearing (grid_line events come incrementally)
  await expect(async () => {
    const cells = await page.locator('.cell').count();
    expect(cells).toBeGreaterThan(10);
  }).toPass({ timeout: 10000 });
});

test.afterAll(async () => {
  await page?.close();
});

test('highlight CSS style tag exists', async () => {
  const styleTag = page.locator('style#nvim-hl-style');
  await expect(styleTag).toBeAttached();
});

test('style tag contains fg and bg color rules', async () => {
  const css = await page.locator('style#nvim-hl-style').textContent();
  expect(css).toMatch(/\.fg-\d+\{color:#[0-9a-f]{6}\}/);
  expect(css).toMatch(/\.bg-\d+\{background-color:#[0-9a-f]{6}\}/);
});

test('rendered cells exist in the grid', async () => {
  // Cells should already be present from beforeAll
  const cells = await page.locator('.cell').count();
  expect(cells).toBeGreaterThan(10);
});

test('cells have text content from test.go', async () => {
  // Grid_line events come incrementally, so poll until content stabilizes
  // Wait for highlighted cells to appear and contain expected content
  await expect(async () => {
    const highlightedCells = await page.locator('.cell[class*="fg-"]').all();
    expect(highlightedCells.length).toBeGreaterThan(10);

    const cellTexts = await Promise.all(highlightedCells.map(c => c.textContent()));
    const allText = cellTexts.join('');

    // Verify known tokens from test.go
    expect(allText).toContain('package');
    expect(allText).toContain('main');
    expect(allText).toContain('import');
    expect(allText).toContain('fmt');
    expect(allText).toContain('func');
    expect(allText).toContain('name');
    expect(allText).toContain('world');
    expect(allText).toContain('Println');
  }).toPass({ timeout: 30000, intervals: [500, 1000, 2000] });
});

test('cells have highlight classes applied', async () => {
  // Highlighted cells should already be present from beforeAll
  const highlightedCells = await page.locator('.cell[class*="fg-"]').count();
  expect(highlightedCells).toBeGreaterThan(10);
});

test('different token types have different fg classes (syntax highlighting works)', async () => {
  // Poll for specific tokens in case content is still settling
  await expect(async () => {
    // Keywords like "package", "import", "func" should have one fg class
    const packageCell = page.locator('.cell').filter({ hasText: /^package$/ }).first();
    await expect(packageCell).toBeVisible();
    const packageClass = await packageCell.getAttribute('class');
    expect(packageClass).toMatch(/fg-\d+/);
    const packageFgClass = packageClass?.match(/fg-\d+/)?.[0];

    // Strings like "fmt" or "world" should have different fg classes
    const stringCell = page.locator('.cell').filter({ hasText: /^"world"$/ }).first();
    await expect(stringCell).toBeVisible();
    const stringClass = await stringCell.getAttribute('class');
    const stringFgClass = stringClass?.match(/fg-\d+/)?.[0];

    // Keywords and strings should have different colors
    expect(packageFgClass).toBeDefined();
    expect(stringFgClass).toBeDefined();
    expect(packageFgClass).not.toBe(stringFgClass);
  }).toPass({ timeout: 15000 });
});

test('every fg class on cells has a matching CSS rule', async () => {
  // Get all unique fg classes from rendered cells
  const cells = await page.locator('.cell[class*="fg-"]').all();
  const fgClasses = new Set<string>();

  for (const cell of cells) {
    const classList = await cell.getAttribute('class');
    const matches = classList?.match(/fg-\d+/g) || [];
    matches.forEach(cls => fgClasses.add(cls));
  }

  expect(fgClasses.size).toBeGreaterThan(0);

  // Get CSS content
  const css = await page.locator('style#nvim-hl-style').textContent();

  // Verify each fg class has a matching CSS rule
  for (const fgClass of fgClasses) {
    const regex = new RegExp(`\\.${fgClass}\\{color:#[0-9a-f]{6}\\}`);
    expect(css).toMatch(regex);
  }
});

test('every bg class on cells has a matching CSS rule', async () => {
  // Get all unique bg classes from rendered cells (may be empty if no backgrounds)
  const cells = await page.locator('.cell[class*="bg-"]').all();
  if (cells.length === 0) {
    // No bg classes is valid - skip this test
    test.skip();
    return;
  }

  const bgClasses = new Set<string>();

  for (const cell of cells) {
    const classList = await cell.getAttribute('class');
    const matches = classList?.match(/bg-\d+/g) || [];
    matches.forEach(cls => bgClasses.add(cls));
  }

  // Get CSS content
  const css = await page.locator('style#nvim-hl-style').textContent();

  // Verify each bg class has a matching CSS rule
  for (const bgClass of bgClasses) {
    const regex = new RegExp(`\\.${bgClass}\\{background-color:#[0-9a-f]{6}\\}`);
    expect(css).toMatch(regex);
  }
});

test('no duplicate fg class definitions in CSS', async () => {
  const css = await page.locator('style#nvim-hl-style').textContent();
  const fgClasses = css?.match(/\.fg-\d+/g) || [];
  const unique = new Set(fgClasses);
  expect(fgClasses.length).toBe(unique.size);
});

test('no duplicate bg class definitions in CSS', async () => {
  const css = await page.locator('style#nvim-hl-style').textContent();
  const bgClasses = css?.match(/\.bg-\d+/g) || [];
  const unique = new Set(bgClasses);
  expect(bgClasses.length).toBe(unique.size);
});
