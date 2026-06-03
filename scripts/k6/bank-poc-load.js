import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    search: {
      executor: 'constant-vus',
      vus: 20,
      duration: '5m',
      exec: 'searchLogs',
    },
    finops: {
      executor: 'constant-vus',
      vus: 10,
      duration: '30s',
      startTime: '5m',
      exec: 'finopsBreakdown',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.001'],
  },
};

const base = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const token = __ENV.K6_TOKEN || __ENV.NEURALOPS_API_TOKEN || '';
const tenant = __ENV.K6_TENANT || 'default';

function headers() {
  const h = { 'X-Tenant-ID': tenant };
  if (token) {
    h.Authorization = `Bearer ${token}`;
  }
  return h;
}

export function searchLogs() {
  const res = http.get(
    `${base}/api/v1/search/logs?q=service:ledger-service&limit=25`,
    { headers: headers() },
  );
  check(res, {
    'search status 200': (r) => r.status === 200,
    'search success envelope': (r) => {
      try {
        return r.json('status') === 'success';
      } catch {
        return false;
      }
    },
  });
  sleep(0.25);
}

export function finopsBreakdown() {
  const res = http.get(
    `${base}/api/v1/finops/costs/breakdown?dimension=team&scope=payments`,
    { headers: headers() },
  );
  check(res, {
    'finops status 200': (r) => r.status === 200,
    'finops success envelope': (r) => {
      try {
        return r.json('status') === 'success';
      } catch {
        return false;
      }
    },
  });
  sleep(0.5);
}
