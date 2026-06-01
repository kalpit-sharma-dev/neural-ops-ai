variable "kubeconfig_path" {
  type        = string
  description = "Path to kubeconfig used by kubernetes/helm providers"
  default     = "~/.kube/config"
}

variable "namespace" {
  type        = string
  description = "Kubernetes namespace for NeuralOps workloads"
  default     = "neuralops"
}

variable "environment" {
  type        = string
  description = "Deployment environment label"
  default     = "production"
}

variable "image_tag" {
  type        = string
  description = "Container image tag for NeuralOps services"
  default     = "latest"
}

variable "gateway_replicas" {
  type        = number
  description = "Initial gateway replica count"
  default     = 2
}

variable "ingress_enabled" {
  type        = bool
  description = "Enable ingress resources"
  default     = true
}

variable "app_host" {
  type        = string
  description = "Frontend ingress hostname"
  default     = "app.neuralops.example"
}

variable "demo_mode" {
  type        = bool
  description = "Enable demo mode in deployed workloads"
  default     = false
}

variable "auth_disabled" {
  type        = bool
  description = "Disable auth on gateway (dev only)"
  default     = false
}
