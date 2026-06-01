import { test, expect } from '@playwright/test';

test.describe('Trace explorer UI', () => {
  test('trace explorer loads', async ({ page }) => {
    await page.goto('/traces');
    await expect(page.getByRole('heading', { name: /trace explorer/i })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole('button', { name: /search/i })).toBeVisible();
  });

  test('can search and open trace detail', async ({ page }) => {
    await page.goto('/traces');
    await page.getByRole('button', { name: /search/i }).click();
    const traceLink = page.locator('a[href*="/traces/trace-demo"]').first();
    await expect(traceLink).toBeVisible({ timeout: 15_000 });
    await traceLink.click();
    await expect(page.getByText(/waterfall|flame graph/i).first()).toBeVisible({ timeout: 15_000 });
  });

  test('trace compare page loads with query params', async ({ page }) => {
    await page.goto('/traces/compare?a=trace-demo-01&b=trace-demo-02');
    await expect(page.getByRole('heading', { name: /compare traces/i })).toBeVisible({ timeout: 15_000 });
  });
});
