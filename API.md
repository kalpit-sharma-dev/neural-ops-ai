# API Reference

Base URL: `/api/v1`

All authenticated endpoints require `Authorization: Bearer <token>` or `X-API-Key`.

## Response envelope

Success:

```json
{
  "status": "success",
  "data": {},
  "timestamp": "2026-05-31T12:00:00Z"
}
```

Error:

```json
{
  "status": "error",
  "errorCode": "AUTH001",
  "message": "invalid api key",
  "timestamp": "2026-05-31T12:00:00Z"
}
```

## Auth

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/config` | Client auth settings |
| POST | `/auth/dev/login` | Dev-only login (non-production) |
| POST | `/auth/oidc/start` | Begin OIDC PKCE |
| GET | `/auth/oidc/callback` | OIDC redirect handler |
| POST | `/auth/oidc/exchange` | Swap one-time code for tokens |
| POST | `/auth/refresh` | Refresh access token |
| GET | `/auth/me` | Current user profile |
| POST | `/auth/logout` | Revoke refresh tokens |

## Ingestion (via gateway)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/logs` | Ingest logs (single or batch) |
| POST | `/metrics` | Ingest metrics |
| POST | `/events` | Ingest deployment/config events |
| POST | `/traces` | Ingest trace spans |
| POST | `/webhooks/{source}` | Alert webhooks |

## Search

| Method | Path | Description |
|--------|------|-------------|
| POST | `/search/logs` | Structured log search |
| POST | `/search/semantic` | Vector / hybrid search |
| POST | `/search/ai` | Natural language search |
| GET | `/search/trace/{traceId}` | Trace log lookup |
| GET | `/search/txn/{txnId}` | Transaction journey |

## Incidents

| Method | Path | Description |
|--------|------|-------------|
| GET | `/incidents` | List incidents |
| GET | `/incidents/{id}` | Incident detail |
| POST | `/incidents/{id}/acknowledge` | Acknowledge |
| POST | `/incidents/{id}/resolve` | Resolve |
| GET | `/incidents/{id}/timeline` | Timeline events |
| GET | `/incidents/{id}/recommendations` | AI recommendations |

## Dashboard & realtime

| Method | Path | Description |
|--------|------|-------------|
| GET | `/dashboard/overview` | Aggregated KPIs |
| GET | `/ws` | WebSocket realtime events |

Swagger UI: `GET /swagger/index.html` on the gateway.
