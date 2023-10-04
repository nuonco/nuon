terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      prefix = "demo-org-"
    }
  }

  required_providers {
    nuon = {
      source  = "nuonco/nuon"
      version = ">= 0.1.1"
    }
  }
}
