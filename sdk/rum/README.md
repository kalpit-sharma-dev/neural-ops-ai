# NeuralOps RUM SDK

Browser SDK for real-user monitoring and session replay.

## Quick start

```html
<script src="https://cdn.neuralops.ai/rum/neuralops-rum.js"></script>
<script>
  NeuralOpsRUM.init({
    endpoint: 'https://your-gateway/api/v1',
    apiKey: 'your-api-key',
    tenantId: 'your-tenant-id',
    replayEnabled: true,
  });
  NeuralOpsRUM.identify('user-123');
</script>
```

## Events

| Endpoint | Purpose |
|----------|---------|
| `POST /api/v1/rum/beacon` | Session metrics (LCP, duration, errors) |
| `POST /api/v1/rum/replay` | Session replay frames (click, navigation, snapshot) |

## Local development

Serve from gateway static or copy `neuralops-rum.js` into `frontend/public/rum/`.
