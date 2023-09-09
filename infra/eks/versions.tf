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
      version = ">= 4.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.4"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.14"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.13.1"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.4.2"
    }
    twingate = {
      source  = "Twingate/twingate"
      version = ">= 0.3.3"
    }
    tfe = {
      source  = "hashicorp/tfe"
      version = ">= 0.36.1"
    }
    utils = {
      source  = "cloudposse/utils"
      version = ">= 0.17.23"
    }
  }
}
