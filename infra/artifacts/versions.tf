terraform {
  required_version = ">= 1.7.5"

  backend "remote" {
    organization = "nuonco"
    workspaces {
      name = "infra-artifacts"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.91.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.17"
    }
    tfe = {
      source  = "hashicorp/tfe"
      version = ">= 0.64.0"
    }
    utils = {
      source  = "cloudposse/utils"
      version = ">= 1.29.0"
    }
  }
}
