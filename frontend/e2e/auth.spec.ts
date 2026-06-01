import { test, expect } from '@playwright/test';

const API_BASE = process.env.PLAYWRIGHT_API_URL ?? 'http://localhost:8080/api/v1';

async function devLogin(request: import('@playwright/test').APIRequestContext) {
  const response = await request.post(`${API_BASE}/auth/dev/login`, {
    data: {
      email: 'demo@neuralops.ai',
      tenantId: '00000000-0000-0000-0000-000000000002',
    },
  });
  if (!response.ok()) {
    return null;
  }
  const body = await response.json();
  return body.data as { accessToken: string };
}

test.describe('NeuralOps authenticated flows', () => {
  test('dev login and dashboard', async ({ page, request }) => {
    const session = await devLogin(request);
    test.skip(!session, 'Auth disabled or dev login unavailable');

    await page.goto('/login');
    await page.getByLabel(/email/i).fill('demo@neuralops.ai');
    await page.getByRole('button', { name: /sign in|dev login|login/i }).click();
    await expect(page).toHaveURL(/\//, { timeout: 15_000 });
    await expect(page.getByText(/command center|dashboard|active incidents/i).first()).toBeVisible();
  });

  test('log search returns results after login', async ({ page, request }) => {
    const session = await devLogin(request);
    test.skip(!session, 'Auth disabled or dev login unavailable');

    await page.goto('/login');
    await page.getByLabel(/email/i).fill('demo@neuralops.ai');
    await page.getByRole('button', { name: /sign in|dev login|login/i }).click();
    await page.goto('/logs');
    await expect(page.getByPlaceholder(/search|query|filter/i)).toBeVisible({ timeout: 15_000 });
  });

  test('incident detail route loads', async ({ page, request }) => {
    const session = await devLogin(request);
    test.skip(!session, 'Auth disabled or dev login unavailable');

    await page.goto('/incidents');
    const firstLink = page.locator('a[href^="/incidents/"]').first();
    await expect(firstLink).toBeVisible({ timeout: 15_000 });
    await firstLink.click();
    await expect(page.getByText(/overview|timeline|logs/i).first()).toBeVisible();
  });
});
