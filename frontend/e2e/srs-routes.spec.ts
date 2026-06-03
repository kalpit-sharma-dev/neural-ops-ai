import { test, expect } from '@playwright/test';

async function devLoginIfNeeded(page: import('@playwright/test').Page) {
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  const emailInput = page.getByLabel(/email/i).first();
  if (!(await emailInput.count())) return;
  await emailInput.fill('demo@neuralops.ai');
  await page.getByRole('button', { name: /dev sign in|sign in|login/i }).click();
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 15_000 });
}

test.describe('SRS phase routes smoke', () => {
  test.beforeEach(async ({ page }) => {
    await devLoginIfNeeded(page);
  });

  for (const path of [
    '/ai-ops',
    '/enterprise-governance',
    '/nfr-certification',
    '/query-workbench',
    '/collectors/fleet',
    '/settings/alert-policies',
    '/settings/alert-suppressions',
    '/observability/materialization',
    '/cloud-finops-network',
    '/business-observability',
  ]) {
    test(`loads ${path}`, async ({ page }) => {
      await page.goto(path, { waitUntil: 'domcontentloaded' });
      await expect(page.locator('main, #main-content').first()).toBeVisible({ timeout: 20_000 });
    });
  }
});
