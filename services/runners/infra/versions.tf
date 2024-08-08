terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      prefix = "runners-"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.61.0"
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
