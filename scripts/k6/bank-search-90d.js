import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 15,
  duration: '3m',
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.01'],
  },
};

const base = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const token = __ENV.K6_TOKEN || __ENV.NEURALOPS_API_TOKEN || '';
const tenant = __ENV.K6_TENANT || 'default';
const retention = __ENV.RETENTION_DAYS || '90';

function headers() {
  const h = { 'X-Tenant-ID': tenant };
  if (token) {
    h.Authorization = `Bearer ${token}`;
  }
  return h;
}

export default function () {
  const res = http.get(
    `${base}/api/v1/search/logs?q=service:ledger-service&limit=50&retentionDays=${retention}`,
    { headers: headers() },
  );
  check(res, {
    '90d search status 200': (r) => r.status === 200,
    '90d search envelope': (r) => {
      try {
        return r.json('status') === 'success';
      } catch {
        return false;
      }
    },
    '90d search p95 under 2s': (r) => r.timings.duration < 2000,
  });
  sleep(0.2);
}
