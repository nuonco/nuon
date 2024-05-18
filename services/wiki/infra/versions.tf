terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      prefix = "wiki-"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
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
