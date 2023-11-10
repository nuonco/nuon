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
      version = ">= 0.3.6"
    }

    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
  }
}
