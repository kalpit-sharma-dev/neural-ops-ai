# Terraform bootstrap for NeuralOps on Kubernetes

Creates the `neuralops` namespace and installs the platform via Helm (`infra/helm/neuralops`).

## Prerequisites

- Terraform >= 1.6
- kubectl access to a target cluster
- Helm 3 (used by Terraform helm provider)

## Usage

```bash
cd infra/terraform
terraform init
terraform plan \
  -var="app_host=app.example.com" \
  -var="api_host=api.example.com" \
  -var="image_tag=0.1.0"
terraform apply
```

## Chart

The Helm chart templates gateway, all backend microservices, HPA/PDB, ConfigMap, and optional ingress.
Override values via Terraform variables or `helm upgrade --values`.

## Notes

- Managed data stores (RDS, MSK, OpenSearch, ElastiCache) should be provisioned in separate Terraform modules and injected into Kubernetes Secrets.
- For local development use `docker compose` and `scripts/quickstart.sh` instead.
