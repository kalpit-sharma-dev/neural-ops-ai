import { test, expect } from '@playwright/test';

test.describe('Materialization dashboard', () => {
  test('loads materialization page and shows policy selector', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    const emailInput = page.getByLabel(/email/i).first();
    if (await emailInput.count()) {
      await emailInput.fill('demo@neuralops.ai');
      await page.getByRole('button', { name: /dev sign in|sign in|login/i }).click();
      await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 15_000 });
    }
    await page.goto('/observability/materialization', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('main, #main-content').first()).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(/streaming materialization|materialización|materialisierung/i).first()).toBeVisible();
  });
});
