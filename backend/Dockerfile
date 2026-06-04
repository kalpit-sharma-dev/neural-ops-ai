# syntax=docker/dockerfile:1

# Compile all Go microservices once; per-service images only copy one binary.
# This avoids 8+ parallel `go build` runs that can crash BuildKit (EOF) on limited hardware.

FROM golang:1.25-alpine AS build-all

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
# Nested module required by root replace before go mod download.
COPY pkg/nexagent/ebpf/go.mod pkg/nexagent/ebpf/go.sum ./pkg/nexagent/ebpf/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -e; \
    for svc in gateway ingestion analysis correlation incident search alerting collector profiler materializer; do \
      echo "==> building ${svc}"; \
      CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o "/out/${svc}" "./cmd/${svc}"; \
    done

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates wget \
    && addgroup -S neuralops \
    && adduser -S neuralops -G neuralops

WORKDIR /app

ARG SERVICE=gateway
COPY --from=build-all /out/${SERVICE} /app/service
COPY --from=build-all /src/configs /app/configs
COPY --from=build-all /src/migrations /app/migrations

USER neuralops

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${HTTP_PORT:-8080}/health || exit 1

ENTRYPOINT ["/app/service"]
