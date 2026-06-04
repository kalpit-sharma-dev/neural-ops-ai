import http from 'k6/http';
import { check, sleep } from 'k6';

const duration = __ENV.K6_INGEST_DURATION || '2m';
const rate = Number(__ENV.K6_INGEST_RATE || 50);

export const options = {
  scenarios: {
    ingest: {
      executor: 'constant-arrival-rate',
      rate,
      timeUnit: '1s',
      duration,
      preAllocatedVUs: 10,
      maxVUs: Math.max(50, rate),
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
const summaryPath = __ENV.K6_SUMMARY_PATH || '';

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

export function handleSummary(data) {
  const failedRate = data.metrics.http_req_failed?.values?.rate ?? 1;
  const p95 = data.metrics.http_req_duration?.values?.['p(95)'] ?? 99999;
  const passed = failedRate < 0.01 && p95 < 3000;
  const report = {
    profile: 'bank-ingest-load',
    generatedAt: new Date().toISOString(),
    baseUrl: base,
    tenant,
    duration,
    rate,
    passed,
    gates: {
      http_req_failed_rate_lt_0_01: failedRate < 0.01,
      http_req_duration_p95_lt_3000ms: p95 < 3000,
    },
    metrics: {
      http_req_failed_rate: failedRate,
      http_req_duration_p95_ms: p95,
      http_reqs: data.metrics.http_reqs?.values?.count ?? 0,
    },
  };
  const out = {};
  if (summaryPath) {
    out[summaryPath] = JSON.stringify(report, null, 2);
  }
  out.stdout = JSON.stringify(report, null, 2);
  return out;
}
