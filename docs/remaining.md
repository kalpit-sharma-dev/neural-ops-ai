# Vendor parity — complete

All long-tail vendor parity items are implemented.

| Area | Deliverable |
|------|-------------|
| **Profiling** | Pyroscope + Grafana Beyla (eBPF) + profiler agent in `docker-compose.yml`; K8s fleet manifest |
| **OneAgent** | deb/rpm build scripts, systemd unit, Docker image, K8s DaemonSet |
| **Mobile** | `@neuralops/mobile-sdk`, store submission guide, EAS production profile |
| **SSO** | Multi-tenant `SSOManager` — per-org IdP from `tenant_policies` |
| **Integrations** | ServiceNow OAuth (instance URL + OAuth flow) |
| **OpenAPI/Pact** | Full `docs/openapi/gateway-v1.yaml` + Pact-style contract tests |
| **Gap matrix** | `UI_DYNATRACE_GAP.md` v1.2 updated |

## Quick links

- Pyroscope UI: http://localhost:4040 (after `quickstart`)
- Build OneAgent deb: `bash deploy/packaging/build-deb.sh`
- K8s fleet: `kubectl apply -f deploy/kubernetes/fleet-profiling.yaml`
- Mobile submit: `mobile/store/SUBMISSION.md`
- OpenAPI test: `make openapi-test`
- **UI improvements (phased):** [UI_IMPROVEMENT_ROADMAP.md](./UI_IMPROVEMENT_ROADMAP.md)

Manual: App Store Connect IDs, Play service account, customer cloud credentials.
