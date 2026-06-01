import { test, expect } from '@playwright/test';

test.describe('NeuralOps smoke tests', () => {
  test('dashboard loads', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('heading', { name: /dashboard|overview|neuralops/i })).toBeVisible({
      timeout: 15_000,
    });
  });

  test('log explorer page loads', async ({ page }) => {
    await page.goto('/logs');
    await expect(page.getByPlaceholder(/search|query|filter/i)).toBeVisible({ timeout: 15_000 });
  });

  test('incidents list loads', async ({ page }) => {
    await page.goto('/incidents');
    await expect(page.getByText(/incident/i).first()).toBeVisible({ timeout: 15_000 });
  });

  test('AI chat page loads', async ({ page }) => {
    await page.goto('/ai-chat');
    await expect(page.getByPlaceholder(/ask|message|question/i)).toBeVisible({ timeout: 15_000 });
  });
});
