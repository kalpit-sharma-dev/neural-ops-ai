import { test, expect } from '@playwright/test';

test.describe('Trace explorer UI', () => {
  test('trace explorer loads', async ({ page }) => {
    await page.goto('/traces');
    await expect(page.getByRole('heading', { name: /trace explorer/i })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole('button', { name: /search/i })).toBeVisible();
  });

  test('can open trace inline and on full page', async ({ page }) => {
    await page.goto('/traces');
    await page.getByRole('button', { name: /search/i }).click();
    const traceRow = page.locator('.trace-row').first();
    await expect(traceRow).toBeVisible({ timeout: 15_000 });
    await traceRow.click();
    await expect(page.getByRole('heading', { name: /trace detail/i })).toBeVisible();
    await expect(page.getByText(/waterfall|flame graph/i).first()).toBeVisible({ timeout: 15_000 });

    await page.getByRole('link', { name: /full page/i }).first().click();
    await expect(page).toHaveURL(/\/traces\/trace-demo/);
    await expect(page.getByText(/waterfall|flame graph/i).first()).toBeVisible({ timeout: 15_000 });
  });

  test('trace compare page loads with query params', async ({ page }) => {
    await page.goto('/traces/compare?a=trace-demo-01&b=trace-demo-02');
    await expect(page.getByRole('heading', { name: /compare traces/i })).toBeVisible({ timeout: 15_000 });
  });
});
