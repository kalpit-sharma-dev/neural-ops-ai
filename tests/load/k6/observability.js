import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    checks: ['rate>0.9'],
  },
};

const BASE = __ENV.API_BASE || 'http://localhost:8080/api/v1';

export default function () {
  const search = http.post(
    `${BASE}/apm/traces/search`,
    JSON.stringify({ size: 20, timeRange: '1h' }),
    { headers: { 'Content-Type': 'application/json', 'X-Tenant-ID': 'default' } },
  );
  check(search, { 'trace search 200': (r) => r.status === 200 });

  const metrics = http.get(`${BASE}/metrics/query?name=latency_p99&service=payment-api`, {
    headers: { 'X-Tenant-ID': 'default' },
  });
  check(metrics, { 'metrics query 200': (r) => r.status === 200 });

  sleep(1);
}
