terraform {
  required_version = ">= 1.3.3"

  backend "remote" {
    organization = "launchpaddev"

    workspaces {
      prefix = "api-gateway-"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }

    tfe = {
      source  = "hashicorp/tfe"
      version = ">= 0.42.0"
    }

    utils = {
      source  = "cloudposse/utils"
      version = ">= 0.17.23"
    }
  }
}
