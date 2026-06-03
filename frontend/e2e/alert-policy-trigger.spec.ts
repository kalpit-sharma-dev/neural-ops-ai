import { test, expect } from '@playwright/test';

test.describe('Alert policy trigger E2E', () => {
  test('policy trigger API returns success envelope', async ({ request }) => {
    const base = process.env.PLAYWRIGHT_API_URL ?? 'http://localhost:8080/api/v1';
    const policies = await request.get(`${base}/alerts/policies`, {
      headers: { 'X-Tenant-ID': 'default' },
    });
    expect(policies.ok()).toBeTruthy();
    const list = (await policies.json()) as { data?: { id: string }[] };
    const policyId = list.data?.[0]?.id ?? 'ap-1';
    const trigger = await request.post(`${base}/alerts/policies/${policyId}/trigger`, {
      headers: { 'X-Tenant-ID': 'default', 'Content-Type': 'application/json' },
      data: { service: 'payment-service', severity: 'P1' },
    });
    expect(trigger.ok()).toBeTruthy();
    const body = await trigger.json();
    expect(body.status).toBe('success');
    expect(body.data).toBeTruthy();
  });
});
