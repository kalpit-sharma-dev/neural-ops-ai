terraform {
  required_providers {
    neuralops = {
      source = "neuralops/neuralops"
    }
  }
}

provider "neuralops" {
  base_url = "http://localhost:8080/api/v1"
}

resource "neuralops_export_job" "daily_warehouse" {
  type        = "warehouse"
  destination = "s3://neuralops-exports/prod/daily"
}

output "export_job_id" {
  value = neuralops_export_job.daily_warehouse.job_id
}
