import { test, expect, type Page } from '@playwright/test';

// Regression test: heading line number flex-shrink
//
// Heading spans contain fill characters (⠀) rendered at the heading's larger
// font-size (2em for H1, 1.5em for H2, etc.). In wide windows, neovim sends
// many fill characters to pad rows to the grid width. At 2x font size these
// fill characters are very wide, causing total heading content to exceed the
// row width. Without flex-shrink:0 on the line number, the flex algorithm
// compresses the line number column, shifting heading text left.
//
// The fix is shrink-0 on LineNumber spans.

let page: Page;

test.describe.configure({ mode: 'serial' });

test.beforeAll(async ({ browser }) => {
  // Use a wide viewport to reproduce the bug — narrow windows don't trigger
  // flex-shrink because there are fewer fill characters.
  page = await browser.newPage({ viewport: { width: 1920, height: 900 } });
  await page.goto('/');

  await expect(page.locator('style#nvim-hl-style')).toBeAttached({ timeout: 15000 });

  // Open the markdown fixture which has h1-h6 headings
  await page.keyboard.type(':e tests/fixtures/table.md');
  await page.keyboard.press('Enter');

  // Wait for markdown content to render
  await expect(async () => {
    const cells = await page.locator('.cell').count();
    expect(cells).toBeGreaterThan(5);
  }).toPass({ timeout: 10000 });

  // Wait for heading to render (treesitter needs time to parse)
  await expect(page.locator('.heading-1').first()).toBeVisible({ timeout: 15000 });
});

test.afterAll(async () => {
  await page?.close();
});

test('headings are rendered with correct classes', async () => {
  await expect(page.locator('.heading-1')).toBeAttached();
  await expect(page.locator('.heading-2')).toBeAttached();
  await expect(page.locator('.heading-3')).toBeAttached();
});

test('all heading line numbers have the same width as normal line numbers', async () => {
  // This is the core regression test. Without shrink-0, H1/H2/H3 line numbers
  // get compressed by flex-shrink due to wide heading fill characters.
  const widths = await page.evaluate(() => {
    const results: { rowIndex: number; lineNumWidth: number; isHeading: boolean }[] = [];
    for (let i = 1; i <= 8; i++) {
      const row = document.getElementById(`row-${i}`);
      if (!row) continue;
      const lineNum = row.children[0] as HTMLElement;
      results.push({
        rowIndex: i,
        lineNumWidth: lineNum?.offsetWidth ?? 0,
        isHeading: !!row.querySelector('.heading'),
      });
    }
    return results;
  });

  // All line numbers should have the same width
  const normalWidth = widths.find(w => !w.isHeading)?.lineNumWidth;
  expect(normalWidth).toBeGreaterThan(0);

  for (const row of widths.filter(w => w.isHeading)) {
    expect(row.lineNumWidth, `heading row ${row.rowIndex} line number width`).toBe(normalWidth);
  }
});

test('heading content x-position matches normal text x-position', async () => {
  const positions = await page.evaluate(() => {
    const results: { rowIndex: number; contentX: number; isHeading: boolean }[] = [];
    for (let i = 1; i <= 8; i++) {
      const row = document.getElementById(`row-${i}`);
      if (!row) continue;
      const heading = row.querySelector('.heading');
      const firstCell = row.querySelector('.cell');
      const contentEl = heading || firstCell;
      if (!contentEl) continue;
      results.push({
        rowIndex: i,
        contentX: contentEl.getBoundingClientRect().x,
        isHeading: !!heading,
      });
    }
    return results;
  });

  const normalX = positions.find(p => !p.isHeading)?.contentX;
  expect(normalX).toBeDefined();

  for (const row of positions.filter(p => p.isHeading)) {
    expect(row.contentX, `heading row ${row.rowIndex} content x-position`).toBe(normalX);
  }
});
