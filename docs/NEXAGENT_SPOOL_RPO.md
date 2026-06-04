# NEXAGENT Air-Gap Spool — RPO Guarantees (GAP-COLL-002)

## Guarantee

| Property | Value |
|----------|-------|
| Max RPO | 300 seconds (configurable via collector) |
| Delivery | At-least-once replay to `POST /api/v1/events` |
| Durability | Disk spool under `NEXAGENT_SPOOL_DIR` |

## Verification

```bash
cd backend && go test ./pkg/nexagent/buffer/... -run TestSpoolReplayDrain -count=1
curl -s http://localhost:8080/api/v1/collectors/spool/guarantee -H 'X-Tenant-ID: default'
```

## Operations

1. Agent enqueues batches when gateway unreachable.
2. On reconnect, `FlushPending` drains spool in FIFO order.
3. Duplicate delivery is idempotent at ingest when `eventId` is present.
