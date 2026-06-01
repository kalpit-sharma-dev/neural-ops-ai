# syntax=docker/dockerfile:1
# Serves static Prometheus metrics for local docker-compose (no K8s API).

FROM nginx:1.27-alpine

RUN rm /etc/nginx/conf.d/default.conf

COPY nginx.conf /etc/nginx/conf.d/kube-metrics.conf
COPY metrics.prom /etc/metrics/metrics.prom

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1
