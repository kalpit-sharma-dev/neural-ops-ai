# Terraform Provider for NeuralOps (v1)

This provider manages enterprise governance resources and export connectors via the NeuralOps Gateway API.

## Requirements

- Terraform >= 1.5
- NeuralOps Gateway reachable at `NEURALOPS_API_URL` (default `http://localhost:8080/api/v1`)
- API token with admin scope (`NEURALOPS_API_TOKEN`)

## Build provider (local dev)

```bash
cd terraform-provider-neuralops
go build -o terraform-provider-neuralops
```

Register for local development:

```bash
export TF_CLI_CONFIG_FILE="$(pwd)/dev.tfrc"
```

See `dev.tfrc` in this directory.

## Example

```hcl
terraform {
  required_providers {
    neuralops = {
      source  = "neuralops/neuralops"
      version = "~> 1.0"
    }
  }
}

provider "neuralops" {
  base_url = "http://localhost:8080/api/v1"
  # token via NEURALOPS_API_TOKEN env
}

resource "neuralops_export_job" "warehouse_daily" {
  type        = "warehouse"
  destination = "s3://neuralops-exports/prod/daily"
}
```

## Resources (v1)

| Resource | API |
|----------|-----|
| `neuralops_export_job` | `POST /exports/{warehouse\|bi\|events}` |
| `neuralops_branding` | `GET/PUT /admin/branding` |
| `neuralops_finops_budget` | `GET/POST/PUT/DELETE /finops/budgets` |
| `neuralops_finops_allocation_rule` | `GET/POST/PUT/DELETE /finops/allocation/rules` |

## Acceptance tests

```bash
cd terraform-provider-neuralops
go test ./... -run TestAcc -count=1
```

Tests skip automatically when `TF_ACC=1` is not set or the gateway is unreachable.
