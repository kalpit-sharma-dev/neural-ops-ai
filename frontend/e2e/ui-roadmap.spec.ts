import { test, expect } from '@playwright/test';

test.describe('UI roadmap smoke', () => {
  test('command palette opens from search', async ({ page }) => {
    await page.goto('/');
    await page.getByLabel('Open command palette').click();
    await expect(page.getByRole('dialog', { name: /command palette/i })).toBeVisible();
    await page.keyboard.press('Escape');
  });

  test('skip link focuses main', async ({ page }) => {
    await page.goto('/');
    await page.keyboard.press('Tab');
    const skip = page.getByRole('link', { name: /skip to main/i });
    if (await skip.isVisible()) {
      await skip.click();
      await expect(page.locator('#main-content')).toBeFocused();
    }
  });

  test('full journey dashboard logs incident', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('heading', { name: /command center/i })).toBeVisible({ timeout: 15_000 });

    await page.goto('/logs');
    await expect(page.getByPlaceholder(/search logs/i)).toBeVisible({ timeout: 15_000 });

    await page.goto('/incidents');
    await expect(page.getByRole('heading', { name: /incidents/i })).toBeVisible({ timeout: 15_000 });
  });

  test('design system forms section', async ({ page }) => {
    await page.goto('/design-system');
    await expect(page.getByRole('heading', { name: /forms/i })).toBeVisible({ timeout: 15_000 });
  });
});
