# GCP & GKE Setup Guide — NeuralOps (zLog Analyzer)

**Audience:** Platform / SRE / FinOps teams deploying NeuralOps to monitor Google Cloud and GKE workloads end-to-end.  
**Version:** 1.0  
**Last updated:** 2026-06-04  

This guide does not skip steps. Follow phases in order unless your organization already completed a phase (e.g. existing GKE fleet).

---

## Table of contents

1. [What you are building](#1-what-you-are-building)
2. [What gets installed in GCP (inventory)](#2-what-gets-installed-in-gcp-inventory)
3. [Roles, access, and ownership](#3-roles-access-and-ownership)
4. [Prerequisites](#4-prerequisites)
5. [Phase 0 — GCP organization & projects](#5-phase-0--gcp-organization--projects)
6. [Phase 1 — Networking & security baseline](#6-phase-1--networking--security-baseline)
7. [Phase 2 — Managed data plane (GCP)](#7-phase-2--managed-data-plane-gcp)
8. [Phase 3 — GKE cluster for NeuralOps (management plane)](#8-phase-3--gke-cluster-for-neuralops-management-plane)
9. [Phase 4 — Deploy NeuralOps platform (Helm)](#9-phase-4--deploy-neuralops-platform-helm)
10. [Phase 5 — GKE workload clusters to monitor](#10-phase-5--gke-workload-clusters-to-monitor)
11. [Phase 6 — Metrics stack on every monitored cluster](#11-phase-6--metrics-stack-on-every-monitored-cluster)
12. [Phase 7 — NeuralOps collector & operator](#12-phase-7--neuralops-collector--operator)
13. [Phase 8 — Logs, traces, and APM (OTEL / Fluent Bit / agents)](#13-phase-8--logs-traces-and-apm-otel--fluent-bit--agents)
14. [Phase 9 — GCP cloud inventory & Cloud Monitoring](#14-phase-9--gcp-cloud-inventory--cloud-monitoring)
15. [Phase 10 — FinOps (GCP billing export)](#15-phase-10--finops-gcp-billing-export)
16. [Phase 11 — Identity, secrets, and hardening](#16-phase-11--identity-secrets-and-hardening)
17. [Phase 12 — UI, tenants, and verification](#17-phase-12--ui-tenants-and-verification)
18. [Ongoing operations](#18-ongoing-operations)
19. [Troubleshooting](#19-troubleshooting)
20. [Current platform limits (read before RFP claims)](#20-current-platform-limits-read-before-rfp-claims)
21. [Related documentation](#21-related-documentation)

---

## 1. What you are building

NeuralOps is **not** a single GCP Marketplace click-to-deploy product today. Production monitoring of your GCP ecosystem uses **three planes**:

| Plane | Where it runs | Purpose |
|-------|---------------|---------|
| **NeuralOps data & control plane** | One GKE cluster (recommended) + managed GCP data services | Gateway, ingest, search, alerting, FinOps APIs, UI backend |
| **Cluster telemetry plane** | Each GKE cluster you monitor | Prometheus (or GMP), kube-state-metrics, node-exporter, OTEL Collector, Fluent Bit |
| **GCP API plane** | GCP APIs + service accounts | Compute inventory, Cloud Monitoring metrics proxy, billing export files |

```mermaid
flowchart TB
  subgraph gcp_org [GCP Organization]
    subgraph mgmt_proj [Project: neuralops-mgmt]
      GKE_MGMT[GKE: neuralops-platform]
      SQL[(Cloud SQL Postgres)]
      MEM[(Memorystore Redis)]
      CH[(ClickHouse Cloud or GKE)]
      KAFKA[(Confluent / self-managed Kafka)]
      ES[(Elastic Cloud / OpenSearch)]
      SEC[Secret Manager]
    end
    subgraph app_proj [Project: workloads-prod]
      GKE_APP[GKE: application clusters]
      GMP[Google Managed Prometheus or in-cluster Prometheus]
      KSM[kube-state-metrics]
      NE[node-exporter]
      FB[Fluent Bit DaemonSet]
      OTEL[OTEL Collector]
    end
    BQ[(BigQuery billing export)]
  end

  Users[Engineers / SRE / Finance] --> ING[Ingress + TLS]
  ING --> GW[NeuralOps Gateway]
  GKE_MGMT --> GW
  GW --> SQL & KAFKA & ES & CH & MEM

  GKE_APP --> GMP
  GMP --> COL[NeuralOps collector Deployment]
  COL --> GW
  FB --> KAFKA
  OTEL --> GW
  GW --> BQ
  GW --> gcp_apis[GCP APIs via SA]
```

**Success criteria:** Engineers use the NeuralOps UI for Kubernetes inventory, logs, traces, alerts, cloud assets, and (optional) FinOps; telemetry flows without demo/seed data in production.

---

## 2. What gets installed in GCP (inventory)

Use this checklist in change management. Adjust names to your standards.

### 2.1 GCP projects (minimum)

| Project ID (example) | Purpose |
|----------------------|---------|
| `neuralops-mgmt-prod` | NeuralOps platform GKE + ingress |
| `neuralops-data-prod` | Cloud SQL, Memorystore, secrets (optional split) |
| `workloads-prod` | Application GKE clusters to monitor |
| `workloads-nonprod` | Staging GKE (same patterns, smaller SKUs) |
| `finops-billing` | BigQuery billing dataset + export SA (can be same as mgmt) |

### 2.2 GCP managed services (production data plane)

| Service | NeuralOps usage | Notes |
|---------|-----------------|-------|
| **Cloud SQL for PostgreSQL 15+** | Tenants, collectors, FinOps, workflows | Multi-AZ, PITR — see [MANAGED_DATA_PLANE.md](./runbooks/MANAGED_DATA_PLANE.md) |
| **Memorystore for Redis 7** | Cache, rate limits | TLS (`rediss://`) |
| **Kafka** | Log/trace pipeline | Confluent Cloud on GCP, or self-managed on GKE |
| **Elasticsearch / OpenSearch** | Log search | Elastic Cloud or managed OpenSearch |
| **ClickHouse** | Transactions / span analytics | ClickHouse Cloud or HA cluster |
| **Cloud Storage** | Backups, FinOps lake (`FINOPS_LAKE_BUCKET`) | Versioning + lifecycle |
| **BigQuery** | Billing export destination | Standard billing export tables |
| **Secret Manager** | DSNs, API keys, OIDC, GCP SA JSON | Via External Secrets Operator |
| **Artifact Registry** | Container images | Mirror `neuralops/*` images |
| **Cloud DNS** | `app.*` / `api.*` records | Point to Ingress LB IP |
| **Certificate Manager** or **cert-manager** | TLS for Ingress | Public or private CA |

### 2.3 GKE — management cluster (NeuralOps platform)

| Workload | Type | Source in repo |
|----------|------|----------------|
| `gateway`, `ingestion`, `analysis`, `correlation`, `incident`, `search`, `alerting` | Deployment (+ HPA) | `infra/helm/neuralops/` |
| `collector` | Deployment | `backend/cmd/collector` — **not in Helm chart yet; deploy manually** |
| `materializer` | Deployment | `backend/cmd/materializer` — compose reference in `infra/docker-compose.yml` |
| `neuralops-collector-operator` | Deployment + CRD | `deploy/kubernetes/operator/` |
| Ingress (NGINX or GKE Gateway) | Ingress | `infra/helm/neuralops/templates/ingress.yaml` |
| Optional: `neuralops-oneagent`, `beyla` | DaemonSet | `deploy/kubernetes/fleet-profiling.yaml` (eBPF profiling) |

### 2.4 GKE — each monitored application cluster

| Workload | Type | Purpose |
|----------|------|---------|
| **kube-state-metrics** | Deployment or Helm chart | K8s object metrics → Prometheus |
| **node-exporter** | DaemonSet | Node CPU/mem/disk |
| **Prometheus** or **Google Managed Prometheus (GMP)** | Managed or in-cluster | Scrape target for NeuralOps `collector` |
| **Fluent Bit** | DaemonSet | Container logs → Kafka |
| **OpenTelemetry Collector** | Deployment or DaemonSet | Traces/metrics → NeuralOps OTLP |
| **NeuralOps collector** (optional per cluster) | Deployment | K8s API sync when running in-cluster with SA |

### 2.5 What NeuralOps does **not** install in your GCP project by default

- VPC, subnets, Cloud NAT, firewall rules (you create via Terraform)
- GKE clusters (you create)
- BigQuery billing export (you enable in Billing console)
- IdP (Okta / Google Workspace / Azure AD) — you configure SSO on gateway

---

## 3. Roles, access, and ownership

| Role | Responsibilities |
|------|------------------|
| **GCP admin** | Projects, VPC, GKE, IAM, billing export, org policies |
| **NeuralOps platform team** | Helm release, secrets, upgrades, collector/operator |
| **Application SRE** | OTEL/Fluent Bit in app namespaces, service names, SLOs |
| **Security** | Workload Identity, mTLS, PII scrubbing, SSO |
| **FinOps** | BigQuery export, invoice reconciliation, `FINOPS_*` files |

### 3.1 IAM roles (reference)

**Service account: `neuralops-platform@...` (management GKE — Workload Identity)**

| GCP role | Why |
|----------|-----|
| `roles/cloudsql.client` | Cloud SQL Auth Proxy / private IP |
| `roles/secretmanager.secretAccessor` | External Secrets |
| `roles/storage.objectViewer` | FinOps lake / billing file buckets (if used) |

**Service account: `neuralops-gcp-inventory@...` (gateway — cloud inventory)**

| GCP role | Why |
|----------|-----|
| `roles/compute.viewer` | GCE aggregated instance list (`cloudinventory` GCP provider) |
| `roles/browser` | Project/folder browse (optional) |

**Service account: `neuralops-billing-reader@...` (export job only)**

| GCP role | Why |
|----------|-----|
| `roles/bigquery.dataViewer` | Read billing export tables (if you automate export to NDJSON) |
| `roles/bigquery.jobUser` | Run export queries |

**Service account: `neuralops-collector@...` (per monitored GKE — K8s API sync)**

| K8s RBAC | Why |
|----------|-----|
| ClusterRole: `get/list/watch` on `pods`, `namespaces`, `deployments` | `K8sAPIClient` in `backend/internal/collector/k8s_api.go` |

Create a **custom** ClusterRole; do not use `cluster-admin`.

---

## 4. Prerequisites

### 4.1 Tools (workstation / CI)

| Tool | Version |
|------|---------|
| `gcloud` | Latest |
| `kubectl` | Within 1 minor of GKE version |
| `helm` | 3.12+ |
| `terraform` | 1.6+ (optional; see `infra/terraform/`) |
| `docker` / `crane` | Image build/push to Artifact Registry |
| `go` | 1.25+ (build collector/operator from source) |
| `openssl` / `curl` | Verification scripts |

### 4.2 NeuralOps artifacts

- Container images for: `gateway`, `ingestion`, `analysis`, `correlation`, `incident`, `search`, `alerting`, `collector`, `materializer`, `collector-operator`
- Helm chart: `infra/helm/neuralops`
- Optional offline license: [bank/AIR_GAP_INSTALL.md](./bank/AIR_GAP_INSTALL.md)

### 4.3 Network

- Private GKE nodes recommended
- Egress to: IdP, LLM (if AI enabled), GCP APIs, managed data plane endpoints
- Ingress: public or internal LB per bank policy

---

## 5. Phase 0 — GCP organization & projects

**Where:** Google Cloud Console → IAM & Admin / Billing, or Terraform.

1. **Link billing account** to all NeuralOps-related projects.
2. **Create projects** (see §2.1). Enable APIs on each:

```bash
export MGMT_PROJECT=neuralops-mgmt-prod
export WORKLOAD_PROJECT=workloads-prod

gcloud config set project "$MGMT_PROJECT"

for API in \
  container.googleapis.com \
  compute.googleapis.com \
  sqladmin.googleapis.com \
  redis.googleapis.com \
  secretmanager.googleapis.com \
  artifactregistry.googleapis.com \
  monitoring.googleapis.com \
  logging.googleapis.com \
  bigquery.googleapis.com \
  iam.googleapis.com \
  dns.googleapis.com; do
  gcloud services enable "$API" --project="$MGMT_PROJECT"
done

gcloud services enable container.googleapis.com --project="$WORKLOAD_PROJECT"
```

3. **Org policies:** Allow Workload Identity, disable public bucket access, enforce uniform bucket-level access.
4. **Labeling standard:** `environment`, `cost-center`, `team`, `neuralops-tenant` (for FinOps tags).

**Exit:** Projects exist, APIs enabled, billing linked.

---

## 6. Phase 1 — Networking & security baseline

**Where:** VPC network in `neuralops-mgmt-prod` and each `workloads-*` project.

1. Create **VPC** with subnets per region (e.g. `us-central1`).
2. **Private Service Connect / private IP** for Cloud SQL and Memorystore.
3. **Cloud NAT** for private nodes pulling images and calling APIs.
4. **Firewall:** default deny; allow internal mesh, health checks, Ingress.
5. **Artifact Registry** repository:

```bash
gcloud artifacts repositories create neuralops \
  --repository-format=docker \
  --location=us-central1 \
  --description="NeuralOps platform images"
```

6. Push images (replace tag with your release):

```bash
export REG=us-central1-docker.pkg.dev/$MGMT_PROJECT/neuralops
docker build -f backend/Dockerfile --build-arg SERVICE=gateway -t $REG/gateway:1.0.0 backend
docker push $REG/gateway:1.0.0
# Repeat for ingestion, analysis, correlation, incident, search, alerting, collector, materializer, collector-operator
```

**Exit:** Nodes can pull from Artifact Registry; DB/cache reachable on private IP.

---

## 7. Phase 2 — Managed data plane (GCP)

**Where:** GCP Console or customer Terraform (outline: `infra/terraform/managed-data-plane/README.md`).  
**Runbook:** [runbooks/MANAGED_DATA_PLANE.md](./runbooks/MANAGED_DATA_PLANE.md)

Provision **HA managed** stores — do not run single-replica Postgres/Kafka in production.

| Store | GCP service | Connection secret key |
|-------|-------------|------------------------|
| Postgres | Cloud SQL | `DATABASE_URL` |
| Redis | Memorystore | `REDIS_URL` (`rediss://`) |
| Kafka | Confluent Cloud or GKE Strimzi | `KAFKA_BROKERS` |
| Elasticsearch | Elastic Cloud | `ELASTICSEARCH_URL` |
| ClickHouse | ClickHouse Cloud | `CLICKHOUSE_DSN` |

### 7.1 Store secrets in Secret Manager

Example (Postgres):

```bash
gcloud secrets create neuralops-database-url \
  --replication-policy=automatic \
  --data-file=- <<EOF
postgres://neuralops:PASSWORD@10.x.x.x:5432/neuralops?sslmode=require
EOF
```

Repeat for `KAFKA_BROKERS`, `ELASTICSEARCH_URL`, `CLICKHOUSE_DSN`, `REDIS_URL`, `NEURALOPS_API_TOKEN`, OIDC client secret, `NEURALOPS_ENCRYPTION_KEY`, JWT keys.

### 7.2 Run database migrations

From a CI job or one-off Job in GKE:

```bash
cd backend
export POSTGRES_DSN="$(gcloud secrets versions access latest --secret=neuralops-database-url)"
go run ./cmd/migrate  # or your Goose migration entrypoint used in CI
```

**Exit:** `psql "$DATABASE_URL" -c 'SELECT 1'` succeeds from a pod in the management VPC.

---

## 8. Phase 3 — GKE cluster for NeuralOps (management plane)

**Where:** GKE → Create cluster in `neuralops-mgmt-prod`.

Recommended settings:

| Setting | Value |
|---------|-------|
| Mode | Standard or Autopilot (Autopilot: validate DaemonSet/eBPF needs) |
| Private nodes | Yes |
| Workload Identity | **Enabled** |
| Release channel | Regular or Stable |
| Node pool | 3+ nodes, `e2-standard-8` minimum for pilot |
| Network policy | Calico or Cilium if using `networkPolicy.enabled` in Helm |

```bash
gcloud container clusters create neuralops-platform \
  --region=us-central1 \
  --enable-private-nodes \
  --enable-ip-alias \
  --workload-pool="${MGMT_PROJECT}.svc.id.goog" \
  --release-channel=regular \
  --num-nodes=3

gcloud container clusters get-credentials neuralops-platform \
  --region=us-central1 \
  --project="$MGMT_PROJECT"
```

Create namespace:

```bash
kubectl create namespace neuralops
kubectl label namespace neuralops istio-injection=enabled  # only if using mesh
```

**Exit:** `kubectl get nodes` healthy.

---

## 9. Phase 4 — Deploy NeuralOps platform (Helm)

**Where:** Workstation CI → management GKE.  
**Chart:** `infra/helm/neuralops`  
**Production values:** `values.yaml` + `values-prod.yaml`

### 9.1 Install External Secrets Operator (recommended)

Map Secret Manager → Kubernetes `Secret` `neuralops-secrets` (see `templates/external-secret.yaml`). Enable in values:

```yaml
externalSecrets:
  enabled: true
  storeRef:
    name: gcp-secret-manager
    kind: ClusterSecretStore
  remoteKey: neuralops/prod
```

Bind GCP SA to K8s SA via Workload Identity:

```bash
gcloud iam service-accounts create neuralops-eso --project="$MGMT_PROJECT"

gcloud secrets add-iam-policy-binding neuralops-database-url \
  --member="serviceAccount:neuralops-eso@${MGMT_PROJECT}.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

gcloud iam service-accounts add-iam-policy-binding \
  neuralops-eso@${MGMT_PROJECT}.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${MGMT_PROJECT}.svc.id.goog[external-secrets/neuralops-eso]"
```

### 9.2 Create `values-gcp-prod.yaml` (customer file — example)

```yaml
global:
  environment: production
  imageRegistry: us-central1-docker.pkg.dev/neuralops-mgmt-prod/neuralops
  imageTag: "1.0.0"   # pin digest in real prod
  imagePullPolicy: IfNotPresent

gateway:
  replicas: 3

config:
  demoMode: false
  authDisabled: false
  authAllowDevLogin: false
  finopsBillingMode: hybrid   # or live when GCP billing file mounted
  finopsRequirePostgres: true
  finopsAllowSimulation: false

auth:
  oidc:
    enabled: true   # Google Workspace / Okta — see SSO runbook

ingress:
  enabled: true
  className: nginx
  hosts:
    app: app.neuralops.yourdomain.com
    api: api.neuralops.yourdomain.com

finopsCur:
  enabled: false   # AWS S3 CUR — for GCP use PVC + FINOPS_GCP_BILLING_FILE (Phase 10)

secrets:
  create: false    # use External Secrets
```

### 9.3 Wire GCP cloud inventory on gateway

Add to ConfigMap or Secret (not in default Helm — extend `templates/configmap.yaml` or patch):

```yaml
NEURALOPS_CLOUD_LIVE: "true"
NEURALOPS_CLOUD_MAX_PAGES: "500"
GCP_PROJECT_ID: "workloads-prod"   # or folder of projects — start with one
GOOGLE_APPLICATION_CREDENTIALS: ""  # prefer Workload Identity — see Phase 11
```

Mount GCP credentials via WI instead of JSON when possible.

### 9.4 Helm install

```bash
cd infra/helm/neuralops
helm upgrade -i neuralops . \
  -n neuralops \
  -f values.yaml \
  -f values-prod.yaml \
  -f values-gcp-prod.yaml \
  --wait --timeout 15m
```

### 9.5 Deploy materializer (Kafka → Postgres)

Not included in default Helm templates. Example Deployment env:

```yaml
KAFKA_BROKERS: "<from secret>"
POSTGRES_DSN: "<from secret>"
MATERIALIZER_CONCURRENCY: "32"
KAFKA_GROUP: neuralops-materializer
```

See [HYPERSCALE.md](./HYPERSCALE.md) and `infra/docker-compose.yml` service `materializer`.

### 9.6 Ingress & DNS

1. Get Ingress IP: `kubectl get ingress -n neuralops`
2. Create Cloud DNS A records for `app.*` and `api.*`
3. Confirm TLS certificate issued (cert-manager or Certificate Manager)

**Exit:**

```bash
curl -fsS https://api.neuralops.yourdomain.com/health
curl -fsS https://api.neuralops.yourdomain.com/ready
```

---

## 10. Phase 5 — GKE workload clusters to monitor

**Where:** `workloads-prod`, `workloads-nonprod`, etc.

1. Create one or more GKE clusters for applications (same networking standards as Phase 3).
2. Register clusters in your fleet (GKE Enterprise / Config Sync) if multi-cluster.
3. Document **which cluster maps to which NeuralOps tenant** (`TENANT_ID` UUID).

**NeuralOps UI:** Infrastructure → Kubernetes / Collectors Fleet — one agent record per cluster.

**Exit:** `kubectl get nodes` on each workload cluster succeeds for platform team.

---

## 11. Phase 6 — Metrics stack on every monitored cluster

**Where:** Each workload GKE cluster — `monitoring` namespace (recommended).

NeuralOps `collector` service reads **Prometheus** (PromQL) and optionally the **Kubernetes API** when running in-cluster.

### 11.1 Install kube-state-metrics

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm upgrade -i kube-state-metrics prometheus-community/kube-state-metrics \
  -n monitoring --create-namespace
```

### 11.2 Install node-exporter

```bash
helm upgrade -i node-exporter prometheus-community/prometheus-node-exporter \
  -n monitoring
```

### 11.3 Prometheus options on GKE

**Option A — Google Managed Prometheus (recommended on GKE)**

- Enable GMP in cluster.
- Configure scrape for kube-state-metrics and node-exporter.
- Expose a query frontend URL to NeuralOps collector (`PROMETHEUS_URL`).

**Option B — In-cluster Prometheus**

```bash
helm upgrade -i prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
```

Note the Prometheus server URL (e.g. `http://prometheus-kube-prometheus-prometheus.monitoring:9090`).

### 11.4 Verify metrics exist

```bash
# Port-forward Prometheus, then:
curl -s 'http://localhost:9090/api/v1/query?query=count(kube_pod_info)' | jq .
curl -s 'http://localhost:9090/api/v1/query?query=count(kube_node_info)' | jq .
```

**Exit:** Both queries return data > 0.

---

## 12. Phase 7 — NeuralOps collector & operator

**Where:** Management cluster **and/or** each workload cluster.

### 12.1 Collector Deployment (required for K8s/host inventory sync)

Build image: `backend/Dockerfile` with `SERVICE=collector`.

Example manifest (`neuralops` namespace on **management** cluster; for multi-cluster, deploy per cluster with distinct `TENANT_ID`):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: neuralops-collector
  namespace: neuralops
spec:
  replicas: 1
  selector:
    matchLabels:
      app: neuralops-collector
  template:
    metadata:
      labels:
        app: neuralops-collector
    spec:
      serviceAccountName: neuralops-collector
      containers:
        - name: collector
          image: us-central1-docker.pkg.dev/MGMT/neuralops/collector:1.0.0
          env:
            - name: POSTGRES_DSN
              valueFrom:
                secretKeyRef:
                  name: neuralops-secrets
                  key: DATABASE_URL
            - name: CLICKHOUSE_DSN
              valueFrom:
                secretKeyRef:
                  name: neuralops-secrets
                  key: CLICKHOUSE_DSN
            - name: PROMETHEUS_URL
              value: "http://prometheus.monitoring.svc:9090"
            - name: TENANT_ID
              value: "00000000-0000-0000-0000-000000000001"
            - name: COLLECTOR_INTERVAL
              value: "60s"
            - name: ALERTING_URL
              value: "http://alerting:8086"
```

**In-cluster K8s API sync:** When `KUBERNETES_SERVICE_HOST` is set, collector prefers live API inventory over PromQL-only mode (`backend/internal/collector/k8s_api.go`).

### 12.2 Collector ServiceAccount RBAC (workload cluster)

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: neuralops-collector
  namespace: neuralops
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: neuralops-collector-read
rules:
  - apiGroups: [""]
    resources: ["pods", "namespaces"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: neuralops-collector-read
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: neuralops-collector-read
subjects:
  - kind: ServiceAccount
    name: neuralops-collector
    namespace: neuralops
```

### 12.3 Collector operator (fleet CRD)

**Where:** Management GKE — `neuralops` namespace.

```bash
kubectl apply -f deploy/kubernetes/operator/crd-collectoragent.yaml
kubectl apply -f deploy/kubernetes/operator/rbac.yaml
kubectl apply -f deploy/kubernetes/operator/deployment.yaml
```

Create API token secret for operator → gateway:

```bash
kubectl create secret generic neuralops-api-token \
  -n neuralops --from-literal=token='YOUR_ADMIN_API_TOKEN'
```

Register agents in UI (**Collectors → Fleet**) or apply `CollectorAgent` CR (see `deployment.yaml` example).

**Exit:** `GET /api/v1/collectors/fleet` lists agents; Postgres table `collector_k8s_clusters` updates (verify via UI Kubernetes view).

---

## 13. Phase 8 — Logs, traces, and APM (OTEL / Fluent Bit / agents)

**Where:** Each workload cluster + application Deployments.  
**Runbook:** [runbooks/BANK_LOG_INGEST.md](./runbooks/BANK_LOG_INGEST.md)  
**Auto-instrumentation:** [bank/AUTO_INSTRUMENTATION.md](./bank/AUTO_INSTRUMENTATION.md), [COLLECTOR_AUTOINSTRUMENTATION.md](./COLLECTOR_AUTOINSTRUMENTATION.md)

### 13.1 Logs — Fluent Bit → Kafka

1. Deploy Fluent Bit DaemonSet (see BANK_LOG_INGEST ConfigMap).
2. Set `KAFKA_BROKERS` to your managed Kafka endpoint.
3. Set `tenant_id` and `service` on every record.
4. Apply **PII scrubbing** before Kafka ([bank/PII_SCRUBBING_GUIDE.md](./bank/PII_SCRUBBING_GUIDE.md)).

### 13.2 Traces & metrics — OpenTelemetry

Application env (example):

```yaml
OTEL_EXPORTER_OTLP_ENDPOINT: "https://api.neuralops.yourdomain.com"
OTEL_SERVICE_NAME: "payment-api"
OTEL_RESOURCE_ATTRIBUTES: "tenant.id=00000000-0000-0000-0000-000000000001"
```

Java: `-javaagent:opentelemetry-javaagent.jar`  
Go/Python/Node: per [COLLECTOR_AUTOINSTRUMENTATION.md](./COLLECTOR_AUTOINSTRUMENTATION.md).

### 13.3 NEXAGENT / eBPF (optional)

- Air-gap spool RPO: [NEXAGENT_SPOOL_RPO.md](./NEXAGENT_SPOOL_RPO.md)
- Kernel matrix check: `bash scripts/verify-ebpf-fleet.sh`
- Optional DaemonSets: `deploy/kubernetes/fleet-profiling.yaml` (Beyla + oneagent profiling)

**GKE note:** eBPF agents require privileged DaemonSets — review Pod Security Standards (likely `privileged` namespace policy).

### 13.4 Gateway ingest endpoints

| Signal | API |
|--------|-----|
| Traces | `POST /api/v1/traces` (OTLP) |
| Logs | Kafka → ingestion → `GET /api/v1/logs/search` |
| APM spans | `POST /api/v1/apm/spans` |

**Exit:** Log search returns real pod logs; trace view shows spans for instrumented services.

---

## 14. Phase 9 — GCP cloud inventory & Cloud Monitoring

**Where:** Gateway Deployment env + GCP IAM.

### 14.1 Enable live inventory

```bash
NEURALOPS_CLOUD_LIVE=true
GCP_PROJECT_ID=workloads-prod
# GOOGLE_APPLICATION_CREDENTIALS=/var/secrets/gcp/key.json  # or Workload Identity
```

Implementation: `backend/internal/cloudinventory/gcp.go` lists **Compute Engine VMs** (aggregated list). Extend IAM if you add more asset types later.

### 14.2 Verify inventory API

```bash
export GATEWAY=https://api.neuralops.yourdomain.com
export TOKEN="<admin JWT>"
export TENANT="00000000-0000-0000-0000-000000000001"

curl -fsS "$GATEWAY/api/v1/cloud/assets?provider=gcp" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT" | jq .
```

Expect `"source": "gcp-sdk"` in tags when live — not demo seed data.

### 14.3 Cloud Monitoring metrics in UI

Cloud metrics handler (`/api/v1/cloud/metrics`) uses, per integration config:

- `prometheusUrl` — query GMP/Prometheus for metric names, **or**
- `monitorProxyUrl` — HTTP proxy you operate that returns NeuralOps `MetricSeries` JSON

Configure integration in UI (**Cloud monitoring → GCP tab**) or store in integration settings with:

```json
{
  "prometheusUrl": "https://gmp-query.example/gmp",
  "monitorProxyUrl": ""
}
```

Without these, UI shows **demo** series (`source: demo` in API).

### 14.4 SNMP / hybrid network (optional)

```bash
NEURALOPS_CLOUD_LIVE=true
SNMP_COMMUNITY=...
SNMP_TARGETS=10.0.0.1,10.0.0.2
SNMP_SITE=dc1
```

---

## 15. Phase 10 — FinOps (GCP billing export)

**Where:** GCP Billing → BigQuery export + NeuralOps gateway PVC/mount.  
**Doc:** [bank/FINOPS_PRODUCTION.md](./bank/FINOPS_PRODUCTION.md)

### 15.1 Enable GCP billing export to BigQuery

1. Console → **Billing** → **Billing export** → BigQuery.
2. Create dataset (e.g. `billing_export`) in `finops-billing` project.
3. Wait 24–48h for first tables (`gcp_billing_export_v1_*`).

### 15.2 Export NDJSON for NeuralOps (daily job)

NeuralOps ingests **NDJSON files** via `FINOPS_GCP_BILLING_FILE` (not live BigQuery API in gateway yet).

Example line format (`backend/internal/finops/connectors/testdata/gcp_billing_sample.ndjson`):

```json
{"provider":"gcp","accountId":"acme-payments","service":"Compute Engine","region":"us-central1","resourceId":"//compute.../instances/pay-1","amortizedCost":312.40,"billingPeriod":"2026-06-01","tags":{"team":"payments","env":"prod"}}
```

**Scheduled query / Cloud Run job (outline):**

1. Query BigQuery export for prior day costs by service/SKU.
2. Write one NDJSON line per resource row to GCS: `gs://neuralops-finops/gcp/YYYY-MM-DD.ndjson`
3. Sync to gateway PVC (CronJob rsync or GCS FUSE).

### 15.3 Mount file & configure gateway

```yaml
FINOPS_BILLING_MODE: live          # or hybrid during migration
FINOPS_REQUIRE_POSTGRES: "true"
FINOPS_ALLOW_SIMULATION: "false"
FINOPS_GCP_BILLING_FILE: /var/lib/neuralops/gcp/billing.ndjson
```

Create PVC + CronJob (mirror AWS CUR pattern in `infra/helm/neuralops/templates/finops-cur-pvc.yaml` but for GCP path).

### 15.4 Run ingest & reconciliation

```bash
curl -X POST "$GATEWAY/api/v1/finops/ingest/run" \
  -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT"

curl -fsS "$GATEWAY/api/v1/finops/reconciliation" \
  -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT" | jq .
```

Gate F target: drift ≤ 1% vs GCP invoice ([bank/GATE_EF_CLOSEOUT_RUNBOOK.md](./bank/GATE_EF_CLOSEOUT_RUNBOOK.md)).

Verify:

```bash
bash scripts/verify-finops-live.sh   # with FINOPS_GCP_BILLING_FILE set
```

---

## 16. Phase 11 — Identity, secrets, and hardening

**Where:** GCP IAM + GKE + Helm values.

### 16.1 Workload Identity (gateway → GCP APIs)

```bash
# K8s SA
kubectl create serviceaccount neuralops-gcp -n neuralops

# GCP SA
gcloud iam service-accounts create neuralops-gcp-inv --project="$MGMT_PROJECT"

gcloud projects add-iam-policy-binding "$WORKLOAD_PROJECT" \
  --member="serviceAccount:neuralops-gcp-inv@${MGMT_PROJECT}.iam.gserviceaccount.com" \
  --role="roles/compute.viewer"

gcloud iam service-accounts add-iam-policy-binding \
  neuralops-gcp-inv@${MGMT_PROJECT}.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${MGMT_PROJECT}.svc.id.goog[neuralops/neuralops-gcp]"

kubectl annotate serviceaccount neuralops-gcp -n neuralops \
  iam.gke.io/gcp-service-account=neuralops-gcp-inv@${MGMT_PROJECT}.iam.gserviceaccount.com
```

Patch gateway `serviceAccountName: neuralops-gcp` and remove JSON key mounts.

### 16.2 SSO (production mandatory)

Gateway rejects unsafe prod config without OIDC/SAML ([production.go](../backend/internal/gateway/config/production.go)):

- [runbooks/SSO_SETUP_BANK.md](./runbooks/SSO_SETUP_BANK.md)
- Google Workspace: OIDC issuer `https://accounts.google.com` or Azure AD for hybrid shops

### 16.3 mTLS (optional east-west)

```bash
bash scripts/verify-mtls.sh
```

### 16.4 Production flags checklist

| Variable | Production value |
|----------|------------------|
| `DEMO_MODE` | `false` |
| `AUTH_DISABLED` | `false` |
| `AUTH_ALLOW_DEV_LOGIN` | `false` |
| `ENVIRONMENT` | `production` |
| `NEURALOPS_CLOUD_LIVE` | `true` |
| `FINOPS_ALLOW_SIMULATION` | `false` (when FinOps live) |

### 16.5 NetworkPolicy

Enable in Helm when CNI supports it:

```yaml
networkPolicy:
  enabled: true
  ingressNamespace: ingress-nginx
```

---

## 17. Phase 12 — UI, tenants, and verification

**Where:** NeuralOps UI `https://app.neuralops.yourdomain.com`

### 17.1 Tenant onboarding

1. Create tenant (admin API or UI **Settings**).
2. Issue API keys for collectors/operator.
3. Map IdP groups → roles (ADMIN, SRE, VIEWER).
4. Set `X-Tenant-ID` on all agent traffic.

### 17.2 UI areas to validate (GCP)

| UI area | What proves GCP setup |
|---------|------------------------|
| **Infrastructure → Kubernetes** | Live pod/deployment counts from collector |
| **Collectors → Fleet** | Registered agents per GKE cluster |
| **Cloud monitoring → GCP** | Assets + metrics (non-demo `source`) |
| **Logs** | Search hits from Fluent Bit pipeline |
| **Traces / APM** | OTEL services appear |
| **FinOps** | GCP line items + reconciliation green |
| **Alerts** | Policies fire on synthetic/K8s signals |

### 17.3 Engineering verification scripts

```bash
# No cluster required
bash scripts/verify-gap-plan.sh

# Stack required
export BASE_URL=https://api.neuralops.yourdomain.com
export NEURALOPS_TENANT=00000000-0000-0000-0000-000000000001
bash scripts/verify-ebpf-fleet.sh
bash scripts/verify-finops-live.sh
bash scripts/verify-materialization.sh   # Kafka + Postgres
```

E2E UI:

```bash
cd frontend && npx playwright test e2e/smoke.spec.ts e2e/ui-actions.spec.ts
```

### 17.4 Gate checklist (production sign-off)

| Gate | GCP-specific evidence |
|------|------------------------|
| **B — Data plane** | Cloud SQL + Memorystore HA, restore drill |
| **C — Ingest** | Real logs/traces from GKE workloads |
| **D — Platform** | Helm prod values, SSO, no demo mode |
| **F — FinOps** | `FINOPS_GCP_BILLING_FILE` + reconciliation ≤ 1% drift |

---

## 18. Ongoing operations

| Task | Frequency | Reference |
|------|-----------|-----------|
| Upgrade NeuralOps images | Per release | `helm upgrade`, migration job |
| Rotate secrets | 90 days | Secret Manager versions |
| Review collector lag | Daily | Prometheus `up`, collector logs |
| Billing file sync | Daily | BigQuery → NDJSON job |
| Kafka lag | Daily | Ingest consumer dashboards |
| Backup restore test | Quarterly | [runbooks/MANAGED_DATA_PLANE.md](./runbooks/MANAGED_DATA_PLANE.md) |
| eBPF kernel upgrades | Per GKE release | `verify-ebpf-fleet.sh` |

---

## 19. Troubleshooting

| Symptom | Likely cause | Action |
|---------|--------------|--------|
| K8s UI shows demo pods | No Prom metrics / no K8s API RBAC | Fix Phase 6–7; check `collector` logs |
| `cloud/assets` empty | `NEURALOPS_CLOUD_LIVE` false or no GCP creds | Phase 14 + WI |
| Cloud metrics flat/demo | Missing `prometheusUrl` / `monitorProxyUrl` | Phase 14.3 |
| No logs in search | Fluent Bit → Kafka misconfigured | Phase 13.1, Kafka ACLs |
| FinOps ingest 0 lines | Missing NDJSON file | Phase 15 |
| Gateway `/ready` fails | Cloud SQL network or secret | Phase 2 + 7 |
| Operator not reconciling | CRD not installed, bad token | Phase 12.3 |

---

## 20. Current platform limits (read before RFP claims)

| Area | Today | Planned / workaround |
|------|-------|----------------------|
| GCP asset inventory | GCE VMs via Compute API | Expand to GKE/Cloud SQL/Run APIs |
| GCP metrics | Prometheus proxy or custom HTTP proxy | Native Cloud Monitoring API adapter |
| GCP billing | NDJSON file ingest | Direct BigQuery connector |
| Helm chart | Core microservices only | Add collector/materializer templates |
| Multi-project | One `GCP_PROJECT_ID` per gateway env | Multiple gateway shards or config rotation |

See [SRS_NEURALOPS_COMPARISON_GAP.md](./SRS_NEURALOPS_COMPARISON_GAP.md) and [GAP_IMPLEMENTATION_STATUS.md](./GAP_IMPLEMENTATION_STATUS.md).

---

## 21. Related documentation

| Document | Topic |
|----------|-------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System design |
| [PRODUCTION_AND_GTM.md](./PRODUCTION_AND_GTM.md) | Deployment models |
| [HYPERSCALE.md](./HYPERSCALE.md) | Operator, cloud SDKs, materializer |
| [runbooks/MANAGED_DATA_PLANE.md](./runbooks/MANAGED_DATA_PLANE.md) | Cloud SQL, Kafka, ES, ClickHouse |
| [runbooks/BANK_LOG_INGEST.md](./runbooks/BANK_LOG_INGEST.md) | Fluent Bit / OTEL |
| [bank/FINOPS_PRODUCTION.md](./bank/FINOPS_PRODUCTION.md) | FinOps env vars |
| [bank/AUTO_INSTRUMENTATION.md](./bank/AUTO_INSTRUMENTATION.md) | Language agents |
| [infra/terraform/README.md](../infra/terraform/README.md) | Helm via Terraform |
| [infra/helm/neuralops/values-prod.yaml](../infra/helm/neuralops/values-prod.yaml) | Production Helm overrides |

---

*End of GCP Setup Guide.*
