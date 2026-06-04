import { test, expect } from '@playwright/test';
import { apiBase, apiHeaders, devLoginIfNeeded } from './helpers/auth';

test.describe.configure({ mode: 'serial' });

function gatewayHealthUrl(): string {
  return apiBase().replace(/\/api\/v1\/?$/, '') + '/health';
}

test.describe('UI action smoke (GAP-UI-002)', () => {
  test.beforeAll(async ({ request }) => {
    const res = await request.get(gatewayHealthUrl(), { timeout: 15_000 });
    expect(res.ok(), `gateway health at ${gatewayHealthUrl()}`).toBeTruthy();
  });

  test.beforeEach(async ({ page }) => {
    await devLoginIfNeeded(page);
  });

  test('Users invite', async ({ page }) => {
    const email = `smoke-${Date.now()}@neuralops.ai`;
    await page.goto('/settings/users', { waitUntil: 'networkidle' });
    await expect(page.getByTestId('users-invite-email')).toBeVisible();
    await page.getByTestId('users-invite-email').fill(email);
    await page.getByTestId('users-invite-btn').click();
    await expect(page.locator(`[data-testid^="users-row-"]`, { hasText: email })).toBeVisible({
      timeout: 15_000,
    });
  });

  test('Users role change', async ({ page }) => {
    await page.goto('/settings/users', { waitUntil: 'networkidle' });
    const row = page.locator('[data-testid^="users-row-"]').first();
    await expect(row).toBeVisible();
    const select = row.locator('select').first();
    const current = await select.inputValue();
    const next = current === 'ADMIN' ? 'SRE' : 'ADMIN';
    await select.selectOption(next);
    await expect(row.locator('.users-row-role')).toContainText(next, { timeout: 10_000 });
  });

  test('Workflow create', async ({ page }) => {
    const name = `Smoke WF ${Date.now()}`;
    await page.goto('/workflows', { waitUntil: 'networkidle' });
    await page.getByTestId('workflows-create-name').fill(name);
    await page.getByTestId('workflows-create-trigger').fill('incident.p2');
    await page.getByTestId('workflows-create-btn').click();
    await expect(page.getByText(name).first()).toBeVisible({ timeout: 15_000 });
  });

  test('Workflow toggle/delete availability', async ({ page }) => {
    await page.goto('/workflows', { waitUntil: 'networkidle' });
    const edit = page.locator('[data-testid^="workflows-edit-"]').first();
    if (!(await edit.isVisible().catch(() => false))) {
      const name = `Smoke WF Actions ${Date.now()}`;
      await page.getByTestId('workflows-create-name').fill(name);
      await page.getByTestId('workflows-create-trigger').fill('incident.p1');
      await page.getByTestId('workflows-create-btn').click();
      await expect(page.locator('[data-testid^="workflows-edit-"]').first()).toBeVisible({
        timeout: 15_000,
      });
    }
    await expect(page.locator('[data-testid^="workflows-toggle-"]').first()).toBeVisible();
    await expect(page.locator('[data-testid^="workflows-delete-"]').first()).toBeVisible();
    await expect(page.locator('[data-testid^="workflows-test-run-"]').first()).toBeVisible();
  });

  test('Marketplace install endpoint', async ({ request }) => {
    const res = await request.post(`${apiBase()}/marketplace/grafana-panels/install`, {
      headers: apiHeaders(),
      data: { config: { grafanaUrl: 'https://grafana.example', apiToken: 'test-token' } },
    });
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.status).toBe('success');
  });

  test('Marketplace page shows card actions', async ({ page }) => {
    await page.goto('/marketplace', { waitUntil: 'networkidle' });
    await expect(page.locator('.marketplace-card__actions').first()).toBeVisible({ timeout: 15_000 });
    const control = page
      .locator('[data-testid^="marketplace-install-"], [data-testid^="marketplace-configure-"]')
      .first();
    await expect(control).toBeVisible({ timeout: 15_000 });
  });

  test('Integrations connect endpoint', async ({ request }) => {
    const res = await request.post(`${apiBase()}/integrations/jira/connect`, {
      headers: apiHeaders(),
      data: { config: { baseUrl: 'https://jira.example', email: 'ops@example.com' } },
    });
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.status).toBe('success');
  });

  test('Integrations page has connect/edit actions', async ({ page }) => {
    await page.goto('/integrations', { waitUntil: 'networkidle' });
    await expect(page.locator('.settings-grid').first()).toBeVisible({ timeout: 15_000 });
    const control = page
      .locator(
        '[data-testid^="integrations-connect-"], [data-testid^="integrations-edit-"], [data-testid^="integrations-disconnect-"]',
      )
      .first();
    await expect(control).toBeVisible({ timeout: 15_000 });
  });

  test('Security list + detail route', async ({ page }) => {
    await page.goto('/security', { waitUntil: 'networkidle' });
    const attackLink = page.locator('[data-testid^="security-attack-link-"]').first();
    const findingLink = page.locator('[data-testid^="security-finding-link-"]').first();
    if (await attackLink.isVisible().catch(() => false)) {
      await attackLink.click();
      await expect(page).toHaveURL(/\/security\/attacks\/.+/);
      return;
    }
    await expect(findingLink).toBeVisible({ timeout: 15_000 });
    await findingLink.click();
    await expect(page).toHaveURL(/\/security\/findings\/.+/);
  });

  test('Settings API key create', async ({ page }) => {
    await page.goto('/settings/api-keys', { waitUntil: 'networkidle' });
    await expect(page.getByTestId('settings-api-key-name')).toBeVisible({ timeout: 15_000 });
    const name = `Smoke Key ${Date.now()}`;
    await page.getByTestId('settings-api-key-name').fill(name);
    await page.getByTestId('settings-api-key-create').click();
    await expect(page.getByText(name).first()).toBeVisible({ timeout: 15_000 });
  });

  test('Settings usage loads', async ({ page }) => {
    await page.goto('/settings/usage', { waitUntil: 'networkidle' });
    await expect(page.getByTestId('usage-logs-ingested')).toBeVisible();
    await expect(page.getByTestId('usage-traces-ingested')).toBeVisible();
    await expect(page.getByTestId('usage-ai-tokens')).toBeVisible();
    await expect(page.getByTestId('usage-active-users')).toBeVisible();
  });
});
