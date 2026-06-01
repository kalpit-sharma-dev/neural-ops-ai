terraform {
  required_version = ">= 1.6.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.30"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.14"
    }
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig_path
}

provider "helm" {
  kubernetes {
    config_path = var.kubeconfig_path
  }
}

resource "kubernetes_namespace" "neuralops" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/part-of" = "neuralops"
      environment                 = var.environment
    }
  }
}

resource "helm_release" "neuralops" {
  name       = "neuralops"
  namespace  = kubernetes_namespace.neuralops.metadata[0].name
  chart      = "${path.module}/../helm/neuralops"
  wait       = true
  timeout    = 600

  values = [
    yamlencode({
      global = {
        namespace   = var.namespace
        environment = var.environment
        imageTag    = var.image_tag
      }
      ingress = {
        enabled = var.ingress_enabled
        hosts = {
          app = var.app_host
          api = var.api_host
        }
      }
      config = {
        demoMode     = var.demo_mode
        authDisabled = var.auth_disabled
      }
    })
  ]
}
