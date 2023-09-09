terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      name = "nuon-dns"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.20.0"
    }

    local = {
      source  = "hashicorp/local"
      version = ">= 2.1.0"
    }

    null = {
      source  = "hashicorp/null"
      version = ">= 3.1.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0.1"
    }
  }
}
