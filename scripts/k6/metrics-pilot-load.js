import http from 'k6/http';
import { check, sleep } from 'k6';

const duration = __ENV.K6_METRICS_DURATION || '1m';
const rate = Number(__ENV.K6_METRICS_RATE || 30);

export const options = {
  scenarios: {
    metrics: {
      executor: 'constant-arrival-rate',
      rate,
      timeUnit: '1s',
      duration,
      preAllocatedVUs: 5,
      maxVUs: 40,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<2000'],
  },
};

const base = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const tenant = __ENV.K6_TENANT || 'default';
const summaryPath = __ENV.K6_SUMMARY_PATH || '';

function headers() {
  return { 'X-Tenant-ID': tenant };
}

export default function () {
  const promRes = http.get(
    `${base}/api/v1/metrics/promql?query=up&range=15m`,
    { headers: headers() },
  );
  check(promRes, { 'promql ok': (r) => r.status === 200 });

  const derivedRes = http.get(`${base}/api/v1/metrics/derived`, { headers: headers() });
  check(derivedRes, { 'derived catalog ok': (r) => r.status === 200 });

  const explainRes = http.post(
    `${base}/api/v1/query/explain`,
    JSON.stringify({ query: 'error rate', from: 'all' }),
    { headers: { ...headers(), 'Content-Type': 'application/json' } },
  );
  check(explainRes, { 'nexql explain ok': (r) => r.status === 200 });
  sleep(0.02);
}

export function handleSummary(data) {
  const failedRate = data.metrics.http_req_failed?.values?.rate ?? 1;
  const p95 = data.metrics.http_req_duration?.values?.['p(95)'] ?? 99999;
  const passed = failedRate < 0.01 && p95 < 2000;
  const report = {
    profile: 'metrics-pilot-load',
    generatedAt: new Date().toISOString(),
    baseUrl: base,
    tenant,
    duration,
    rate,
    passed,
    gates: {
      http_req_failed_rate_lt_0_01: failedRate < 0.01,
      http_req_duration_p95_lt_2000ms: p95 < 2000,
    },
    metrics: {
      http_req_failed_rate: failedRate,
      http_req_duration_p95_ms: p95,
    },
  };
  const out = {};
  if (summaryPath) {
    out[summaryPath] = JSON.stringify(report, null, 2);
  }
  out.stdout = JSON.stringify(report, null, 2);
  return out;
}
