/**
 * NeuralOps OneAgent-style auto-instrumentation for Node.js.
 * Patches http/https and fetch to emit OTLP-compatible spans to the ingest endpoint.
 */
const http = require('http');
const https = require('https');

const DEFAULTS = {
  serviceName: process.env.NEURALOPS_SERVICE || 'node-service',
  ingestUrl: process.env.NEURALOPS_INGEST_URL || 'http://localhost:8080/api/v1/apm/spans',
  tenantId: process.env.NEURALOPS_TENANT_ID || 'default',
  sampleRate: Number(process.env.NEURALOPS_SAMPLE_RATE || '1'),
};

function shouldSample(rate) {
  return Math.random() < rate;
}

function emitSpan(span) {
  const body = JSON.stringify({ spans: [span] });
  const url = new URL(DEFAULTS.ingestUrl);
  const lib = url.protocol === 'https:' ? https : http;
  const req = lib.request(
    {
      hostname: url.hostname,
      port: url.port || (url.protocol === 'https:' ? 443 : 80),
      path: url.pathname + url.search,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(body),
        'X-Tenant-ID': DEFAULTS.tenantId,
      },
    },
    (res) => res.resume(),
  );
  req.on('error', () => {});
  req.write(body);
  req.end();
}

function wrapRequest(orig, protocol) {
  return function patchedRequest(options, callback) {
    const start = Date.now();
    const req = orig.call(this, options, callback);
    const host = typeof options === 'string' ? options : options.host || options.hostname || 'unknown';
    const path = typeof options === 'string' ? '/' : options.path || '/';
    req.on('response', (res) => {
      if (!shouldSample(DEFAULTS.sampleRate)) return;
      emitSpan({
        traceId: cryptoRandomHex(32),
        spanId: cryptoRandomHex(16),
        serviceName: DEFAULTS.serviceName,
        operationName: `${protocol.toUpperCase()} ${path}`,
        startTime: new Date(start).toISOString(),
        durationMs: Date.now() - start,
        tags: { host, statusCode: String(res.statusCode), protocol },
      });
    });
    return req;
  };
}

function cryptoRandomHex(len) {
  const bytes = require('crypto').randomBytes(len / 2);
  return bytes.toString('hex');
}

function init(opts = {}) {
  Object.assign(DEFAULTS, opts);
  if (!http.__neuralopsPatched) {
    http.request = wrapRequest(http.request, 'http');
    https.request = wrapRequest(https.request, 'https');
    http.__neuralopsPatched = true;
  }
  if (globalThis.fetch && !globalThis.fetch.__neuralopsPatched) {
    const origFetch = globalThis.fetch;
    globalThis.fetch = async function neuralopsFetch(input, init) {
      const start = Date.now();
      const url = typeof input === 'string' ? input : input.url;
      const res = await origFetch(input, init);
      if (shouldSample(DEFAULTS.sampleRate)) {
        emitSpan({
          traceId: cryptoRandomHex(32),
          spanId: cryptoRandomHex(16),
          serviceName: DEFAULTS.serviceName,
          operationName: `FETCH ${url}`,
          startTime: new Date(start).toISOString(),
          durationMs: Date.now() - start,
          tags: { statusCode: String(res.status), protocol: 'fetch' },
        });
      }
      return res;
    };
    globalThis.fetch.__neuralopsPatched = true;
  }
  return DEFAULTS;
}

module.exports = { init, emitSpan, DEFAULTS };
