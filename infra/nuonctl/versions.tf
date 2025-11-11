terraform {
  required_version = ">= 1.7.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.91.0"
    }
  }

  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "nuonco"

    workspaces {
      name = "infra-nuonctl"
    }
  }
}
