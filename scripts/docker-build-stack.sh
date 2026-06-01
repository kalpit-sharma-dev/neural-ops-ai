#!/usr/bin/env bash
# Builds all compose images (fresh by default). Does not start containers.
# Usage: QUICKSTART_USE_CACHE=1 bash scripts/docker-build-stack.sh  # faster
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/infra/docker-compose.yml"
BACKEND_DIR="${ROOT_DIR}/backend"
BACKEND_BUILD_TAG="${BACKEND_BUILD_TAG:-neuralops-backend-build:local}"

log() {
  printf '[docker-build] %s\n' "$*"
}

if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose -f "$COMPOSE_FILE")
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE=(docker-compose -f "$COMPOSE_FILE")
else
  log "Docker Compose is not installed"
  exit 1
fi

PARALLEL="${COMPOSE_PARALLEL_LIMIT:-2}"
export COMPOSE_PARALLEL_LIMIT="$PARALLEL"
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

BUILD_EXTRA=()
COMPOSE_BUILD_EXTRA=()
if [[ "${QUICKSTART_USE_CACHE:-0}" != "1" ]]; then
  BUILD_EXTRA=(--no-cache)
  COMPOSE_BUILD_EXTRA=(--no-cache)
fi

"${COMPOSE[@]}" down --remove-orphans --rmi local 2>/dev/null || true
docker rmi -f "${BACKEND_BUILD_TAG}" 2>/dev/null || true
docker image prune -f >/dev/null 2>&1 || true

log "Step 1/2: Shared Go compile (all backend services, one pass)..."
docker build \
  "${BUILD_EXTRA[@]}" \
  --file "${BACKEND_DIR}/Dockerfile" \
  --target build-all \
  --tag "${BACKEND_BUILD_TAG}" \
  "${BACKEND_DIR}"

log "Step 2/2: Compose images (parallel limit=${PARALLEL})..."
"${COMPOSE[@]}" build "${COMPOSE_BUILD_EXTRA[@]}"

docker image prune -f >/dev/null 2>&1 || true

log "Done. Start with: docker compose -f infra/docker-compose.yml up -d --force-recreate"
