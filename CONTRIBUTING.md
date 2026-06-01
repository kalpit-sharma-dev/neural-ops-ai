# Contributing

## Development setup

1. Install Go 1.22+, Node 20+, Docker
2. Clone the repository
3. Start infrastructure: `docker compose -f infra/docker-compose.yml up -d`
4. Backend: `cd backend && go build ./...`
5. Frontend: `cd frontend && npm install && npm run dev`

## Code standards

- Follow existing package layout under `backend/internal/`
- Use structured logging (zap) and context propagation
- Add migrations for schema changes (Postgres + ClickHouse when analytics-related)
- Prefer tenant-scoped queries and indexes

## Testing

```bash
# Unit tests
cd backend && go test ./...

# Integration tests (Docker required)
cd backend && go test -tags=integration ./tests/integration/...

# Frontend build
cd frontend && npm run build
```

## Pull requests

- Keep changes focused and production-ready
- Update `CHANGELOG.md` for user-visible changes
- Ensure CI passes: lint, build, tests

## Security

- Never commit secrets, API keys, or private keys
- Use `secret.example.yaml` / environment variables for credentials
