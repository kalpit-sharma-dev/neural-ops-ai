import { test, expect } from '@playwright/test';

test.describe('Materialization pipeline API', () => {
  test('policy trigger produces score buckets when materializer runs', async ({ request }) => {
    test.skip(!process.env.PLAYWRIGHT_API_URL && !process.env.CI, 'requires live stack (compose-smoke sets PLAYWRIGHT_BASE_URL)');

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

    let found = false;
    for (let i = 0; i < 12; i++) {
      const scores = await request.get(`${base}/alerts/policies/${policyId}/scores`, {
        headers: { 'X-Tenant-ID': 'default' },
      });
      if (scores.ok()) {
        const body = (await scores.json()) as { data?: unknown[] };
        if ((body.data?.length ?? 0) > 0) {
          found = true;
          break;
        }
      }
      await new Promise((r) => setTimeout(r, 5000));
    }
    expect(found).toBeTruthy();
  });
});
