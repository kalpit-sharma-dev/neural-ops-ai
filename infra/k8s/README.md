# NeuralOps Kubernetes manifests

Apply the platform stack to a Kubernetes cluster:

```bash
kubectl apply -k infra/k8s
```

Before applying:

1. Copy `secret.example.yaml` to `secret.yaml`, fill credentials, and add `secret.yaml` to `kustomization.yaml`.
2. Update hostnames in `frontend-ingress.yaml` and `configmap.yaml`.
3. Build and push service images (`neuralops/gateway`, `neuralops/ingestion`, etc.) from `backend/Dockerfile`.

Manifests include:

- Namespace, ConfigMap, example Secret
- Gateway Deployment/Service/HPA/PDB
- Backend microservice Deployments/Services
- Frontend Deployment/Service
- Ingress for UI and API

Data plane components (Postgres, Kafka, Elasticsearch, Redis, ClickHouse, Qdrant, Jaeger) are expected as cluster services or external managed services referenced in `configmap.yaml`.
