terraform {
  required_version = ">= 1.3.6"

  backend "remote" {
    organization = "launchpaddev"

    workspaces {
      prefix = "workers-orgs-"
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
