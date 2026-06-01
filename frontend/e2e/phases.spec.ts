import { test, expect } from '@playwright/test';

/**
 * Phase coverage E2E — one happy path per Dynatrace implementation plan phase.
 */
test.describe('Phase 1 — Alerting', () => {
  test('alerts page and rules tab', async ({ page }) => {
    await page.goto('/alerts');
    await expect(page.getByRole('heading', { name: /alerts/i })).toBeVisible({ timeout: 15_000 });
    await page.getByRole('button', { name: 'Rules' }).click();
    await expect(page.getByText(/create rule/i)).toBeVisible();
  });
});

test.describe('Phase 2 — APM', () => {
  test('trace search and detail', async ({ page }) => {
    await page.goto('/traces');
    await page.getByRole('button', { name: /search/i }).click();
    await page.locator('a[href*="/traces/"]').first().click({ timeout: 15_000 });
    await expect(page.getByRole('button', { name: /flame graph/i })).toBeVisible();
  });
});

test.describe('Phase 3 — Metrics & dashboards', () => {
  test('metrics and dashboards load', async ({ page }) => {
    await page.goto('/metrics');
    await expect(page.getByRole('heading', { name: /metrics explorer/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/dashboards');
    await expect(page.getByRole('heading', { name: /dashboards/i })).toBeVisible();
  });
});

test.describe('Phase 4 — Smartscape', () => {
  test('service map loads', async ({ page }) => {
    await page.goto('/service-map');
    await expect(page.getByText(/service map|topology/i).first()).toBeVisible({ timeout: 15_000 });
  });
});

test.describe('Phase 5 — Admin', () => {
  test('settings users and api keys', async ({ page }) => {
    await page.goto('/settings/users');
    await expect(page.getByRole('heading', { name: /users/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/settings/api-keys');
    await expect(page.getByRole('heading', { name: /api keys/i })).toBeVisible();
  });
});

test.describe('Phase 6 — Logs', () => {
  test('log settings', async ({ page }) => {
    await page.goto('/logs/settings');
    await expect(page.getByRole('heading', { name: /log/i })).toBeVisible({ timeout: 15_000 });
  });
});

test.describe('Phase 7 — SLOs & anomalies', () => {
  test('slos and anomalies', async ({ page }) => {
    await page.goto('/slos');
    await expect(page.getByRole('heading', { name: /slos/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/anomalies');
    await expect(page.getByRole('heading', { name: /anomaly/i })).toBeVisible();
  });
});

test.describe('Phase 8 — Infrastructure', () => {
  test('infra and kubernetes', async ({ page }) => {
    await page.goto('/infrastructure');
    await expect(page.getByRole('heading', { name: /infrastructure/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/kubernetes');
    await expect(page.getByRole('heading', { name: /kubernetes/i })).toBeVisible();
  });
});

test.describe('Phase 9 — Database & middleware', () => {
  test('databases and middleware', async ({ page }) => {
    await page.goto('/databases');
    await expect(page.getByRole('heading', { name: /database/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/middleware');
    await expect(page.getByRole('heading', { name: /middleware/i })).toBeVisible();
  });
});

test.describe('Phase 10 — RUM & synthetic', () => {
  test('rum and synthetic monitors', async ({ page }) => {
    await page.goto('/rum');
    await expect(page.getByRole('heading', { name: /real user monitoring/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/synthetic');
    await expect(page.getByRole('heading', { name: /synthetic/i })).toBeVisible();
  });
});

test.describe('Phase 11 — Workflows & notebooks', () => {
  test('workflows and notebooks', async ({ page }) => {
    await page.goto('/workflows');
    await expect(page.getByRole('heading', { name: /workflow/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/notebooks');
    await expect(page.getByRole('heading', { name: /notebook/i })).toBeVisible();
  });
});

test.describe('Phase 12 — Security & integrations', () => {
  test('security and integrations', async ({ page }) => {
    await page.goto('/security');
    await expect(page.getByRole('heading', { name: /security/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/integrations');
    await expect(page.getByRole('heading', { name: /integrations/i })).toBeVisible();
  });
});
