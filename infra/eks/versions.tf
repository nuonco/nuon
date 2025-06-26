terraform {
  required_version = ">= 1.7.5"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      prefix = "infra-eks-"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.73.0, < 6.0.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.16.1, < 3.0.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.14"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.33.0"
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
      version = ">= 0.59.0"
    }
    utils = {
      source  = "cloudposse/utils"
      version = ">= 1.24.0"
    }
  }
}
