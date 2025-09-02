terraform {
  required_version = ">= 1.7.5"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      prefix = "infra-orgs-"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.94.1"
    }

    tfe = {
      source  = "hashicorp/tfe"
      version = ">= 0.36.1"
    }
  }
}
