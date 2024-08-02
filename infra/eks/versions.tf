terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      prefix = "infra-eks-"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.61.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.14"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.14"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.31.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.6.2"
    }
    twingate = {
      source  = "Twingate/twingate"
      version = ">= 3.0.8"
    }
    tfe = {
      source  = "hashicorp/tfe"
      version = ">= 0.57.1"
    }
    utils = {
      source  = "cloudposse/utils"
      version = ">= 1.24.0"
    }
  }
}
