terraform {
  required_version = ">= 1.7.5"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      name = "sandboxes"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.27"
    }
  }
}
