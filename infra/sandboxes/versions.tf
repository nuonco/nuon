terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      name = "sandboxes"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
  }
}
