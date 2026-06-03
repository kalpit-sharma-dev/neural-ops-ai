# East-West mTLS & Service Mesh (Wave 1.3)

**ID:** PROD-NET-01

## Objective

Encrypt and authenticate **service-to-service** traffic inside the Kubernetes cluster. North-south TLS terminates at ingress; east-west protects against lateral movement if a pod is compromised.

## Options

| Approach | Best for | Notes |
|----------|----------|-------|
| **Istio / ASM** | Full L7 policy, bank mesh standards | PeerAuthentication `STRICT` |
| **Linkerd** | Lightweight mTLS | Automatic, less config |
| **Cilium** | eBPF + wire encryption | Good with NexAgent NPM pilot |
| **App-level mTLS** | Minimal mesh | Demo: `infra/mtls-proxy` + `scripts/verify-mtls.sh` |

## Recommended (bank self-hosted): Istio STRICT mTLS

### 1. Install Istio (customer cluster)

```bash
istioctl install --set profile=default -y
kubectl label namespace neuralops istio-injection=enabled --overwrite
```

### 2. Helm — enable mesh values

```bash
helm upgrade -i neuralops infra/helm/neuralops \
  -f infra/helm/neuralops/values.yaml \
  -f infra/helm/neuralops/values-prod.yaml \
  --set mesh.enabled=true \
  --set mesh.mtlsMode=STRICT
```

Chart renders `PeerAuthentication` when `mesh.enabled=true` (namespace-wide STRICT).

### 3. Verify

```bash
# Sidecars injected
kubectl get pods -n neuralops -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].name}{"\n"}{end}'

# No plaintext between services (Istio metrics)
istioctl authn tls-check $(kubectl get pod -n neuralops -l app=gateway -o jsonpath='{.items[0].metadata.name}').neuralops
```

### 4. Compose / dev (no mesh)

```bash
bash scripts/gen-mtls-certs.sh
docker compose -f infra/docker-compose.yml up -d mtls-proxy
bash scripts/verify-mtls.sh
```

## Gateway → backend

Until all services present sidecars:

- Keep cluster DNS names (`http://ingestion:8081`) **inside** mesh-enabled namespace only.
- Deny ingress to backend ports via `NetworkPolicy` — only gateway ServiceAccount may reach backends.

## Certificate rotation

| Layer | Rotation |
|-------|----------|
| Ingress TLS | cert-manager, 90-day |
| Mesh CA | Istio CA or customer PKI — annual |
| Client certs (NexAgent) | Per-agent cert, 30–90 day |

## Exit criteria (Gate B)

- [ ] `mesh.enabled=true` in staging OR documented customer mesh equivalent
- [ ] `verify-mtls.sh` passes in CI (compose profile)
- [ ] NetworkPolicy blocks direct backend access from default namespace

## Related

- [infra/certs/README.md](../../infra/certs/README.md)
- [SECURITY_PACK.md](../bank/SECURITY_PACK.md)
