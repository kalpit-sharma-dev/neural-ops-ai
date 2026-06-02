import { test, expect } from '@playwright/test';

async function devLoginIfNeeded(page: import('@playwright/test').Page) {
  await page.goto('/login', { waitUntil: 'domcontentloaded' });
  const emailInput = page.getByLabel(/email/i).first();
  if (!(await emailInput.count())) return;

  await emailInput.fill('demo@neuralops.ai');
  await page.getByRole('button', { name: /dev sign in|sign in|login/i }).click();
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 15_000 });

  const cookieBannerAccept = page.getByRole('button', { name: /accept/i }).first();
  if (await cookieBannerAccept.count()) {
    await cookieBannerAccept.click().catch(() => {});
  }
}

test.describe('NeuralOps smoke tests', () => {
  test.beforeEach(async ({ page }) => {
    await devLoginIfNeeded(page);
  });

  test('dashboard loads', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByText(/command center/i).first()).toBeVisible({ timeout: 15_000 });
  });

  test('log explorer page loads', async ({ page }) => {
    await page.goto('/logs');
    await expect(page.getByRole('heading', { name: 'Log Explorer' }).first()).toBeVisible({ timeout: 15_000 });
  });

  test('incidents list loads', async ({ page }) => {
    await page.goto('/incidents');
    await expect(page.getByText(/incident/i).first()).toBeVisible({ timeout: 15_000 });
  });

  test('AI chat page loads', async ({ page }) => {
    await page.goto('/ai-chat');
    await expect(page.getByPlaceholder(/ask|message|question/i)).toBeVisible({ timeout: 15_000 });
  });

  test.describe('critical UI actions', () => {
    test('settings users invite flow exposes stable hooks', async ({ page }) => {
      await page.goto('/settings/users', { waitUntil: 'networkidle' });
      await expect(page.getByTestId('users-invite-btn')).toBeVisible();
      await expect(page.locator('[data-testid^="users-row-"]').first()).toBeVisible();
    });

    test('workflows action controls expose stable hooks', async ({ page }) => {
      await page.goto('/workflows', { waitUntil: 'networkidle' });
      await expect(page.locator('[data-testid^="workflows-test-run-"]').first()).toBeVisible();
      await expect(page.locator('[data-testid^="workflows-toggle-"]').first()).toBeVisible();
      await expect(page.locator('[data-testid^="workflows-edit-"]').first()).toBeVisible();
      await expect(page.locator('[data-testid^="workflows-delete-"]').first()).toBeVisible();
    });

    test('marketplace install/configure controls expose stable hooks', async ({ page }) => {
      await page.goto('/marketplace', { waitUntil: 'networkidle' });
      const control = page
        .locator('[data-testid^="marketplace-install-"], [data-testid^="marketplace-configure-"]')
        .first();
      await expect(control).toBeVisible();
    });

    test('integrations connect/edit/disconnect controls expose stable hooks', async ({ page }) => {
      await page.goto('/integrations', { waitUntil: 'networkidle' });
      const control = page
        .locator(
          '[data-testid^="integrations-connect-"], [data-testid^="integrations-edit-"], [data-testid^="integrations-disconnect-"]',
        )
        .first();
      await expect(control).toBeVisible();
    });

    test('security attack links expose stable hooks and open detail', async ({ page }) => {
      await page.goto('/security', { waitUntil: 'networkidle' });
      const attackLink = page.locator('[data-testid^="security-attack-link-"]').first();
      await expect(attackLink).toBeVisible();
      await attackLink.click();
      await expect(page).toHaveURL(/\/security\/attacks\/.+/);
    });

    test('settings usage counters expose stable hooks', async ({ page }) => {
      await page.goto('/settings/usage', { waitUntil: 'networkidle' });
      await expect(page.getByTestId('usage-logs-ingested')).toBeVisible();
      await expect(page.getByTestId('usage-traces-ingested')).toBeVisible();
      await expect(page.getByTestId('usage-ai-tokens')).toBeVisible();
      await expect(page.getByTestId('usage-active-users')).toBeVisible();
    });
  });
});
