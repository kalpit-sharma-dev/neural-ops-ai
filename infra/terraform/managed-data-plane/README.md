# Managed Data Plane — Terraform Contract (Wave 1.4)

This folder documents **connection inputs** for NeuralOps Helm. Provision RDS/MSK/OpenSearch/ClickHouse/Redis in the **customer** Terraform root module; do not commit credentials here.

## Variables to export (customer module)

```hcl
variable "neuralops_namespace" {
  type    = string
  default = "neuralops"
}

# Pass to External Secrets or Sealed Secrets — never plain text in Helm values
output "database_url" {
  value     = "postgres://${var.db_user}:@${var.db_endpoint}:5432/neuralops?sslmode=require"
  sensitive = true
}

output "kafka_brokers" {
  value = join(",", var.msk_bootstrap_brokers)
}

output "elasticsearch_url" {
  value = "https://${var.opensearch_endpoint}:443"
}

output "clickhouse_dsn" {
  value     = "clickhouse://${var.ch_host}:9000/neuralops"
  sensitive = true
}

output "redis_url" {
  value = "rediss://${var.redis_primary}:6379"
}
```

## Example AWS resources (outline)

| Resource | Terraform resource type |
|----------|-------------------------|
| Postgres | `aws_db_instance` or `aws_rds_cluster` |
| Kafka | `aws_msk_cluster` |
| Search | `aws_opensearch_domain` |
| Redis | `aws_elasticache_replication_group` |
| ClickHouse | Customer VM/K8s operator or ClickHouse Cloud endpoint |

## Link to application Helm

```hcl
module "neuralops_app" {
  source = "../"
  # ...
}

# External Secrets Operator maps SM secrets → kubernetes secret neuralops-secrets
```

## Validation

After apply:

```bash
kubectl exec -n neuralops deploy/gateway -- wget -qO- http://localhost:8080/ready
psql "$DATABASE_URL" -c 'SELECT 1'
```

See [MANAGED_DATA_PLANE.md](../../../docs/runbooks/MANAGED_DATA_PLANE.md).
