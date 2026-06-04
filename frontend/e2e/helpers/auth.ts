import type { Page } from '@playwright/test';

/** Dev login when the login form is shown (local / compose with dev auth). */
export async function devLoginIfNeeded(page: Page) {
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

export function apiBase(): string {
  return (process.env.PLAYWRIGHT_API_URL ?? 'http://localhost:8080/api/v1').replace(/\/$/, '');
}

export function apiHeaders(): Record<string, string> {
  return {
    'Content-Type': 'application/json',
    'X-Tenant-ID': process.env.PLAYWRIGHT_TENANT ?? 'default',
  };
}
