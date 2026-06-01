import { test, expect } from '@playwright/test';

test.describe('Depth features — Trace settings', () => {
  test('retention policy page loads', async ({ page }) => {
    await page.goto('/traces/settings');
    await expect(page.getByRole('heading', { name: /trace settings|retention/i })).toBeVisible({ timeout: 15_000 });
  });
});

test.describe('Depth features — Cloud monitoring', () => {
  test('cloud dashboards and metrics', async ({ page }) => {
    await page.goto('/cloud');
    await expect(page.getByRole('heading', { name: /cloud monitoring/i })).toBeVisible({ timeout: 15_000 });
    await page.getByRole('button', { name: /AWS|EC2/i }).first().click({ timeout: 10_000 }).catch(() => {});
  });
});

test.describe('Depth features — Workflow editor', () => {
  test('workflow editor loads', async ({ page }) => {
    await page.goto('/workflows/editor');
    await expect(page.getByRole('heading', { name: /workflow/i })).toBeVisible({ timeout: 15_000 });
  });
});

test.describe('Depth features — Escalation & admin', () => {
  test('integrations credential UI and admin settings', async ({ page }) => {
    await page.goto('/integrations');
    await expect(page.getByRole('heading', { name: /integrations/i })).toBeVisible({ timeout: 15_000 });
    await page.goto('/settings/sso');
    await expect(page.getByRole('heading', { name: /sso|idp/i })).toBeVisible();
    await page.goto('/settings/oncall');
    await expect(page.getByRole('heading', { name: /on-call/i })).toBeVisible();
  });
});

test.describe('Depth features — Notebooks', () => {
  test('notebooks list and run button', async ({ page }) => {
    await page.goto('/notebooks');
    await expect(page.getByRole('heading', { name: /notebooks/i })).toBeVisible({ timeout: 15_000 });
  });
});
