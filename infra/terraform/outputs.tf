output "namespace" {
  value       = kubernetes_namespace.neuralops.metadata[0].name
  description = "Deployed Kubernetes namespace"
}

output "helm_release" {
  value       = helm_release.neuralops.name
  description = "Helm release name"
}

output "helm_chart_path" {
  value       = "${path.module}/../helm/neuralops"
  description = "Local Helm chart path"
}
