import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    ingest: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 10,
      maxVUs: 50,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<3000'],
  },
};

const base = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const token = __ENV.K6_TOKEN || __ENV.NEURALOPS_API_TOKEN || '';
const tenant = __ENV.K6_TENANT || 'default';

function headers() {
  const h = { 'Content-Type': 'application/json', 'X-Tenant-ID': tenant };
  if (token) {
    h.Authorization = `Bearer ${token}`;
  }
  return h;
}

export default function () {
  const ts = new Date().toISOString();
  const logBody = JSON.stringify({
    timestamp: ts,
    level: 'INFO',
    service: 'ledger-service',
    message: 'bank ingest load test event',
    traceId: `trace-load-${__VU}-${__ITER}`,
    txnId: `txn-load-${__VU}`,
  });
  const logRes = http.post(`${base}/api/v1/logs`, logBody, { headers: headers() });
  check(logRes, { 'log ingest accepted': (r) => r.status === 200 || r.status === 202 });

  const traceBody = JSON.stringify({
    resourceSpans: [{
      resource: { attributes: [{ key: 'service.name', value: { stringValue: 'payment-service' } }] },
      scopeSpans: [{
        spans: [{
          traceId: '00112233445566778899001122334455',
          spanId: '0011223344556677',
          name: 'POST /payments',
          startTimeUnixNano: `${Date.now()}000000`,
          endTimeUnixNano: `${Date.now() + 120}000000`,
          status: { code: 1 },
        }],
      }],
    }],
  });
  const traceRes = http.post(`${base}/api/v1/traces`, traceBody, { headers: headers() });
  check(traceRes, { 'trace ingest accepted': (r) => r.status === 200 || r.status === 202 });
  sleep(0.05);
}
