#!/usr/bin/env bash
# run-ui-actions-smoke.sh — GAP-UI-002: run 11 Playwright UI action checks and emit JSON summary.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND="${ROOT}/frontend"
OUT="${FRONTEND}/ui-smoke-actions-resume.json"
BASE_URL="${PLAYWRIGHT_BASE_URL:-http://localhost:3000}"
API_URL="${PLAYWRIGHT_API_URL:-http://localhost:8080/api/v1}"

cd "${FRONTEND}"
export PLAYWRIGHT_BASE_URL="${BASE_URL}"
export PLAYWRIGHT_API_URL="${API_URL}"

echo "==> UI action smoke (base=${BASE_URL}, api=${API_URL})"

HEALTH_URL="${API_URL%/api/v1}/health"
if ! curl -fsS "${HEALTH_URL}" >/dev/null 2>&1; then
  echo "ERROR: gateway not reachable at ${HEALTH_URL}" >&2
  echo "Start stack: docker compose -f infra/docker-compose.yml up -d" >&2
  exit 1
fi

npx playwright test e2e/ui-actions.spec.ts --reporter=json > "${OUT}.raw.json" || true

node - "${OUT}.raw.json" "${OUT}" <<'NODE'
const fs = require('fs');
const [rawPath, outPath] = process.argv.slice(2);
let pass = 0;
let fail = 0;
const report = [];
try {
  const raw = JSON.parse(fs.readFileSync(rawPath, 'utf8'));
  const suites = raw.suites ?? [];
  const walk = (suite, prefix = '') => {
    const title = prefix ? `${prefix} > ${suite.title}` : suite.title;
    for (const spec of suite.specs ?? []) {
      const name = spec.title;
      const ok = (spec.tests ?? []).every((t) =>
        (t.results ?? []).every((r) => r.status === 'passed' || r.status === 'skipped'),
      );
      if (ok) {
        pass++;
        report.push({ name, status: 'PASS', issues: [] });
      } else {
        fail++;
        const issues = (spec.tests ?? [])
          .flatMap((t) => t.results ?? [])
          .filter((r) => r.status !== 'passed' && r.status !== 'skipped')
          .map((r) => r.error?.message ?? `status=${r.status}`);
        report.push({ name, status: 'FAIL', issues: issues.length ? issues : ['test failed'] });
      }
    }
    for (const child of suite.suites ?? []) walk(child, title);
  };
  for (const s of suites) walk(s);
} catch (e) {
  report.push({ name: 'playwright runner', status: 'FAIL', issues: [String(e)] });
  fail = 1;
}
const summary = { pass, fail, warn: 0, total: pass + fail };
fs.writeFileSync(outPath, JSON.stringify({ summary, report }, null, 2));
console.log(JSON.stringify(summary));
process.exit(fail > 0 ? 1 : 0);
NODE

rm -f "${OUT}.raw.json"
echo "Report: ${OUT}"
