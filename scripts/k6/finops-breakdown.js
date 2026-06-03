import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<2000'],
  },
};

const base = __ENV.BASE_URL || 'http://localhost:8080';

export default function finopsBreakdown() {
  const res = http.get(`${base}/api/v1/finops/costs/breakdown?dimension=team`, {
    headers: { 'X-Tenant-ID': 'default' },
  });
  check(res, {
    'status is 200': (r) => r.status === 200,
    'success envelope': (r) => r.json('status') === 'success',
  });
  sleep(0.5);
}
