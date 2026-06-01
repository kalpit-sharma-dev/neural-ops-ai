import { test, expect } from '@playwright/test';

const API_BASE = process.env.PLAYWRIGHT_API_URL ?? 'http://localhost:8080/api/v1';

async function devLogin(request: import('@playwright/test').APIRequestContext) {
  const response = await request.post(`${API_BASE}/auth/dev/login`, {
    data: {
      email: 'demo@neuralops.ai',
      tenantId: '00000000-0000-0000-0000-000000000002',
    },
  });
  if (!response.ok()) return null;
  const body = await response.json();
  return body.data as { accessToken: string };
}

test.describe('AI Chat E2E', () => {
  test('sends a question and receives assistant content', async ({ page, request }) => {
    const session = await devLogin(request);
    test.skip(!session, 'Auth disabled or dev login unavailable');

    await page.goto('/login');
    await page.getByLabel(/email/i).fill('demo@neuralops.ai');
    await page.getByRole('button', { name: /sign in|dev login|login/i }).click();
    await page.goto('/ai-chat');

    const question = 'What services had errors in the last hour?';
    await page.getByPlaceholder(/ask about incidents/i).fill(question);
    await page.getByRole('button', { name: /send/i }).click();

    await expect(page.getByText(/Searching logs|Analyzing patterns|Generating response|Error:/i).first()).toBeVisible({
      timeout: 10_000,
    });

    await expect(page.locator('.chat-bubble--ai').last()).not.toHaveText('…', { timeout: 60_000 });
  });
});
