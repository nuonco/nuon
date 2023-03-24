terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "launchpaddev"

    workspaces {
      name = "powertools"
    }
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "> 4.20.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "> 4.20.0"
    }
  }
}

