# NeuralOps OneAgent (Node.js)

Lightweight auto-instrumentation for HTTP/HTTPS and `fetch`.

```javascript
const { init } = require('@neuralops/oneagent');
init({ serviceName: 'checkout-api', ingestUrl: 'http://gateway:8080/api/v1/apm/spans' });
```

Environment variables:

- `NEURALOPS_SERVICE` — service name
- `NEURALOPS_INGEST_URL` — span ingest endpoint
- `NEURALOPS_TENANT_ID` — tenant header
- `NEURALOPS_SAMPLE_RATE` — 0–1 head sampling
