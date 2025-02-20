terraform {
  required_version = ">= 1.7.5"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      prefix = "infra-clickhouse-"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.67.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.14"
    }
    tfe = {
      source  = "hashicorp/tfe"
      version = ">= 0.36.1"
    }
    utils = {
      source  = "cloudposse/utils"
      version = ">= 0.17.23"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.6.2"
    }
  }
}
