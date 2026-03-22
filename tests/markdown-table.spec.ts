import { test, expect, type Page } from '@playwright/test';

// These tests validate the markdown table rendering pipeline:
// 1. Neovim starts, test opens table.md via vim command
// 2. Lua treesitter detects pipe_table → rpcnotify → Go caches TableMeta
// 3. optimizeGrid annotates rows with TableRowOpts (tableId, rowType, cells)
// 4. Frontend groups rows by tableId and renders <table> elements

let page: Page;

test.describe.configure({ mode: 'serial' });

test.beforeAll(async ({ browser }) => {
  page = await browser.newPage();
  await page.goto('/');

  await expect(page.locator('style#nvim-hl-style')).toBeAttached({ timeout: 15000 });

  // Open the markdown table fixture
  await page.keyboard.type(':e tests/fixtures/table.md');
  await page.keyboard.press('Enter');

  // Wait for content to appear
  await expect(async () => {
    const cells = await page.locator('.cell').count();
    expect(cells).toBeGreaterThan(5);
  }).toPass({ timeout: 10000 });

  // Wait for treesitter to parse and table metadata to arrive via rpcnotify
  await expect(async () => {
    const tables = await page.locator('table').count();
    expect(tables).toBeGreaterThan(0);
  }).toPass({ timeout: 15000, intervals: [500, 1000, 2000] });
});

test.afterAll(async () => {
  await page?.close();
});

test('table element exists in the DOM', async () => {
  const tables = await page.locator('table').count();
  expect(tables).toBeGreaterThan(0);
});

test('table has thead with header row', async () => {
  const thead = page.locator('table thead');
  await expect(thead).toBeAttached();
  const headerCells = await thead.locator('th').count();
  expect(headerCells).toBe(3);
});

test('table header contains correct text', async () => {
  const headerTexts = await page.locator('table thead th').allTextContents();
  expect(headerTexts.map(t => t.trim())).toEqual(['Name', 'Age', 'City']);
});

test('table has tbody with data rows', async () => {
  const tbody = page.locator('table tbody');
  await expect(tbody).toBeAttached();
  const dataRows = await tbody.locator('tr').count();
  expect(dataRows).toBe(3);
});

test('table data contains correct text', async () => {
  const rows = await page.locator('table tbody tr').all();
  expect(rows.length).toBe(3);

  const row1 = await rows[0].locator('td').allTextContents();
  expect(row1.map(t => t.trim())).toEqual(['Alice', '30', 'New York']);

  const row2 = await rows[1].locator('td').allTextContents();
  expect(row2.map(t => t.trim())).toEqual(['Bob', '25', 'London']);

  const row3 = await rows[2].locator('td').allTextContents();
  expect(row3.map(t => t.trim())).toEqual(['Carol', '35', 'Tokyo']);
});

test('separator row is not visible as a table row', async () => {
  // The separator/border rows should not appear as data in the rendered table
  const allText = await page.locator('table').textContent();
  // Should not contain raw separator dashes or box-drawing borders
  expect(allText).not.toContain('---');
  expect(allText).not.toContain('┌');
  expect(allText).not.toContain('└');
});

test('cells contain span elements with highlight classes', async () => {
  const cellSpans = await page.locator('table .cell').count();
  expect(cellSpans).toBeGreaterThan(0);
});
