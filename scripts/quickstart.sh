#!/usr/bin/env bash
# Full stack quickstart: tears down old compose images, rebuilds without cache (default),
# recreates containers, and starts NeuralOps.
#
# Faster repeat runs (use Docker layer cache):
#   QUICKSTART_USE_CACHE=1 bash scripts/quickstart.sh
#
# See docs/LOCAL_TEST_URLS.md for test URLs after startup.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/infra/docker-compose.yml"
BACKEND_DIR="${ROOT_DIR}/backend"
BACKEND_BUILD_TAG="${BACKEND_BUILD_TAG:-neuralops-backend-build:local}"

log() {
  printf '[quickstart] %s\n' "$*"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    log "Missing required command: $1"
    exit 1
  fi
}

wait_for_url() {
  local url="$1"
  local name="$2"
  local attempts="${3:-60}"

  for ((i = 1; i <= attempts; i++)); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      log "$name is healthy"
      return 0
    fi
    sleep 2
  done

  log "Timed out waiting for $name at $url"
  return 1
}

log "Checking prerequisites..."
require_cmd docker
require_cmd curl

if [[ ! -f "${ROOT_DIR}/infra/certs/gateway-client.crt" ]]; then
  if command -v openssl >/dev/null 2>&1; then
    log "Generating dev mTLS certificates..."
    bash "${ROOT_DIR}/scripts/gen-mtls-certs.sh"
  else
    log "openssl not found; skipping mTLS cert generation (gateway may fail if INTERNAL_MTLS_ENABLED=true)"
  fi
fi

if command -v go >/dev/null 2>&1; then
  HAS_GO=true
else
  HAS_GO=false
  log "Go not found; seed script will be skipped (install Go to seed demo data locally)"
fi

if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose -f "$COMPOSE_FILE")
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE=(docker-compose -f "$COMPOSE_FILE")
else
  log "Docker Compose is not installed"
  exit 1
fi

export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
export COMPOSE_PARALLEL_LIMIT="${COMPOSE_PARALLEL_LIMIT:-2}"

BUILD_EXTRA=()
COMPOSE_BUILD_EXTRA=()
if [[ "${QUICKSTART_USE_CACHE:-0}" != "1" ]]; then
  BUILD_EXTRA=(--no-cache)
  COMPOSE_BUILD_EXTRA=(--no-cache)
  log "Fresh build mode (no Docker cache). Set QUICKSTART_USE_CACHE=1 to speed up repeat runs."
else
  log "Cache-enabled build (QUICKSTART_USE_CACHE=1)."
fi

log "Stopping stack and removing compose-built images..."
"${COMPOSE[@]}" down --remove-orphans --rmi local 2>/dev/null || true

log "Removing prior backend build artifact image (if present)..."
docker rmi -f "${BACKEND_BUILD_TAG}" 2>/dev/null || true

log "Pruning dangling images from previous builds..."
docker image prune -f >/dev/null 2>&1 || true

log "Building images (shared Go compile + compose services)..."
log "  Step 1/2: compiling all Go services..."
docker build \
  "${BUILD_EXTRA[@]}" \
  --file "${BACKEND_DIR}/Dockerfile" \
  --target build-all \
  --tag "${BACKEND_BUILD_TAG}" \
  "${BACKEND_DIR}"

log "  Step 2/2: building all compose images (max ${COMPOSE_PARALLEL_LIMIT} parallel)..."
"${COMPOSE[@]}" build "${COMPOSE_BUILD_EXTRA[@]}"

log "Pruning unused dangling layers after build..."
docker image prune -f >/dev/null 2>&1 || true

log "Starting NeuralOps stack (force-recreate containers from new images)..."
"${COMPOSE[@]}" up -d --force-recreate --remove-orphans

log "Waiting for core services..."
wait_for_url "http://localhost:8080/health" "Gateway"
wait_for_url "http://localhost:9200" "Elasticsearch"
wait_for_url "http://localhost:8123/ping" "ClickHouse"
wait_for_url "http://localhost:3000/" "Frontend"

if [[ "$HAS_GO" == "true" ]]; then
  log "Seeding demo data (50k logs, 1,747 transactions, 100 deployments)..."
  log "  This may take several minutes on first run."
  (
    cd "$BACKEND_DIR"
    POSTGRES_DSN="${POSTGRES_DSN:-postgres://neuralops:neuralops@localhost:5432/neuralops?sslmode=disable}" \
    CLICKHOUSE_DSN="${CLICKHOUSE_DSN:-clickhouse://default:@localhost:9000/neuralops}" \
    ELASTICSEARCH_URL="${ELASTICSEARCH_URL:-http://localhost:9200}" \
    INGESTION_URL="${INGESTION_URL:-http://localhost:8081/api/v1/logs}" \
      go run ./scripts/seed
  ) || log "Seed script failed (dependencies may still be starting); retry with: cd backend && go run ./scripts/seed"
  log "  Export artifacts: backend/scripts/seed/output/"
  log "  Fast seed (5k logs): make seed-fast"
fi

log ""
log "NeuralOps is running."
log "  Test URLs: docs/LOCAL_TEST_URLS.md"
log "  UI:       http://localhost:3000  (hard-refresh: Ctrl+Shift+R if styles look stale)"
log "  Gateway:  http://localhost:8080/health"
log "  Info:     http://localhost:8080/api/v1/info"
log "  Grafana:  http://localhost:3001 (admin/admin)"
log "  Prometheus: http://localhost:9090"
log "  Jaeger:   http://localhost:16686"
log "  Pyroscope: http://localhost:4040 (continuous profiling)"
log "  Beyla eBPF + OneAgent + profiler agents run in compose"
log "  Collector: syncs K8s/synthetic monitors from Prometheus → Postgres (service: collector)"
log "  Keycloak: http://localhost:8088 (admin/admin)"
log "  Metrics:  http://localhost:8080/metrics"
log ""
log "Auth:"
log "  Dev login: demo@neuralops.ai (tenant 00000000-0000-0000-0000-000000000002)"
log "  OIDC IdP:  Keycloak realm neuralops (client neuralops-ui)"
log "  Keycloak user password: demo1234"
log ""
if [[ -f "${ROOT_DIR}/infra/certs/gateway-client.crt" ]]; then
  log "Verifying Keycloak OIDC..."
  bash "${ROOT_DIR}/scripts/verify-keycloak-oidc.sh" || log "OIDC verification failed (Keycloak may still be starting)"
  log "Verifying SAML SP metadata..."
  bash "${ROOT_DIR}/scripts/verify-saml-metadata.sh" || log "SAML metadata verification failed"
  log "Verifying mTLS proxy..."
  bash "${ROOT_DIR}/scripts/verify-mtls.sh" || log "mTLS verification failed"
fi
log ""
log "Stop with: docker compose -f infra/docker-compose.yml down"
