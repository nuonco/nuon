terraform {
  required_version = ">= 1.3.7"

  backend "remote" {
    organization = "nuonco"

    workspaces {
      name = "infra-vercel"
    }
  }

  required_providers {
    vercel = {
      source  = "vercel/vercel"
      version = "~> 1.8"
    }
  }
}
