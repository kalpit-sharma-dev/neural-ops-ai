# Managed data plane module (Wave 1.4)
#
# Wire customer-managed Postgres, Kafka, Elasticsearch, ClickHouse, and Redis.
# This module documents required endpoints; provision via your cloud vendor or
# existing Terraform modules (RDS, MSK, OpenSearch, ClickHouse Cloud, ElastiCache).

terraform {
  required_version = ">= 1.5.0"
}

variable "environment" {
  type        = string
  description = "staging or production"
}

variable "postgres_dsn" {
  type        = string
  description = "Postgres connection string for NeuralOps metadata"
  sensitive   = true
}

variable "kafka_brokers" {
  type        = string
  description = "Comma-separated Kafka bootstrap brokers"
}

variable "elasticsearch_url" {
  type        = string
  description = "Elasticsearch/OpenSearch HTTPS endpoint"
}

variable "clickhouse_dsn" {
  type        = string
  description = "ClickHouse native or HTTP DSN"
  sensitive   = true
}

variable "redis_url" {
  type        = string
  description = "Redis URL for rate limiting and quotas"
  sensitive   = true
}

output "helm_values_snippet" {
  description = "Paste into values-staging.yaml / values-prod.yaml externalSecrets data keys"
  value = {
    POSTGRES_DSN      = var.postgres_dsn
    KAFKA_BROKERS     = var.kafka_brokers
    ELASTICSEARCH_URL = var.elasticsearch_url
    CLICKHOUSE_DSN    = var.clickhouse_dsn
    REDIS_URL         = var.redis_url
    ENVIRONMENT       = var.environment
  }
  sensitive = true
}
