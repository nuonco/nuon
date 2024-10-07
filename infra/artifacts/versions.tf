terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"
    workspaces {
      name = "infra-artifacts"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.67.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.14"
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
