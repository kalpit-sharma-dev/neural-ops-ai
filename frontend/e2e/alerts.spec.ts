import { test, expect } from '@playwright/test';

test.describe('Alerts UI', () => {
  test('alerts page loads with tabs', async ({ page }) => {
    await page.goto('/alerts');
    await expect(page.getByRole('heading', { name: /alerts/i })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole('button', { name: 'Rules' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Channels' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'History' })).toBeVisible();
  });

  test('can switch to rules tab', async ({ page }) => {
    await page.goto('/alerts');
    await page.getByRole('button', { name: 'Rules' }).click();
    await expect(page.getByText(/create rule/i)).toBeVisible({ timeout: 10_000 });
  });
});
