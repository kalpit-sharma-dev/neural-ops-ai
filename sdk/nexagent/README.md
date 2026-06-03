# NEXAGENT

Unified NeuralOps telemetry agent scaffold:

- **eBPF / OTLP** pipelines (integrate with existing `sdk/oneagent`)
- **Air-gapped buffer** via `backend/pkg/nexagent/buffer` disk spool + replay
- **Fleet control** via collector operator + gateway fleet API

```bash
cd backend
go build -o nexagent ./cmd/nexagent
NEXAGENT_BUFFER_DIR=/var/lib/nexagent ./nexagent -replay
```

Environment: `NEXAGENT_GATEWAY`, `NEXAGENT_BUFFER_DIR`.
