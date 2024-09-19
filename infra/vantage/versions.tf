terraform {
  required_version = ">= 1.3.7"


  backend "remote" {
    organization = "nuonco"

    workspaces {
      name = "infra-vantage"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.67.0"
    }
    tfe = {
      source  = "hashicorp/tfe"
      version = ">= 0.36.1"
    }
    utils = {
      source  = "cloudposse/utils"
      version = ">= 0.17.23"
    }
    vantage = {
      source = "vantage-sh/vantage"
      version = "0.1.24"
    }
  }
}
