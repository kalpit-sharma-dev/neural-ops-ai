# Changelog

All notable changes to this project are documented here.

## [0.1.0] - 2026-05-31

### Added

- Full microservices stack: gateway, ingestion, analysis, correlation, incident, search, alerting
- React dashboard with Log Explorer, Incident Detail, Service Map, AI Chat, Transaction Journey
- Phase 20 demo seed: 50k logs, 500 transactions, 100 deployments, UPI outage scenario
- OIDC PKCE login with secure one-time exchange codes
- Tenant isolation for Postgres, Elasticsearch, Qdrant, and gateway auth
- Auth audit events and ClickHouse audit replication
- OpenTelemetry tracing to Jaeger, Prometheus metrics, 5 Grafana dashboards
- Kubernetes manifests and Terraform namespace bootstrap
- Integration tests (Postgres, Elasticsearch, auth flows, seed volumes)
- Enterprise documentation set

### Changed

- Docker Compose enables auth with dev login for local demos
- Elasticsearch seed loads per-tenant indexes

### Security

- Logout revokes refresh tokens
- Auth endpoint rate limiting
- Subscription status enforced for API keys and OIDC users
