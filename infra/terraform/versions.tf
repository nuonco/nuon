terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "launchpaddev"

    workspaces {
      name = "infra-terraform"
    }
  }

  required_providers {
    tfe = {
      version = "~> 0.46.0"
    }
  }
}
