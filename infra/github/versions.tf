terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "launchpaddev"

    workspaces {
      name = "infra-github"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.28.0"
    }

    github = {
      source  = "integrations/github"
      version = "> 4.30.0"
    }
  }
}
