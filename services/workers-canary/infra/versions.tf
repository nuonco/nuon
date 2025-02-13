terraform {
  required_version = ">= 1.7.5"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      prefix = "workers-canary-"
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
  }
}
