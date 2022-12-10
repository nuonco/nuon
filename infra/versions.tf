terraform {
  required_version = ">= 1.3.3"

  backend "remote" {
    organization = "launchpaddev"

    workspaces {
      prefix = "api-"
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
