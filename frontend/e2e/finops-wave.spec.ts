import { test, expect } from '@playwright/test';

async function devLoginIfNeeded(page: import('@playwright/test').Page) {
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  const emailInput = page.getByLabel(/email/i).first();
  if (!(await emailInput.count())) return;
  await emailInput.fill('demo@neuralops.ai');
  await page.getByRole('button', { name: /dev sign in|sign in|login/i }).click();
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 15_000 });
}

const FINOPS_SUBTABS = [
  'Overview',
  'Allocation',
  'Anomalies',
  'Optimization',
  'Budgets',
  'Commitments',
  'Sustainability',
  'Chargeback',
  'What-If',
  'Reports',
] as const;

test.describe('FinOps module sub-tabs', () => {
  test.beforeEach(async ({ page }) => {
    await devLoginIfNeeded(page);
    await page.goto('/cloud-finops-network', { waitUntil: 'domcontentloaded' });
    await page.getByRole('button', { name: /^FinOps$/i }).click();
  });

  for (const subTab of FINOPS_SUBTABS) {
    test(`renders ${subTab} sub-tab`, async ({ page }) => {
      await page.getByTestId(`finops-subtab-${subTab.toLowerCase()}`).click();
      await expect(page.locator('main, #main-content').first()).toBeVisible({ timeout: 20_000 });
    });
  }
});
