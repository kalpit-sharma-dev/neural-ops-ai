# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE=gateway
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/service ./cmd/${SERVICE}

FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget \
    && addgroup -S neuralops \
    && adduser -S neuralops -G neuralops

WORKDIR /app

COPY --from=builder /bin/service /app/service
COPY --from=builder /src/configs /app/configs
COPY --from=builder /src/migrations /app/migrations

USER neuralops

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${HTTP_PORT:-8080}/health || exit 1

ENTRYPOINT ["/app/service"]
