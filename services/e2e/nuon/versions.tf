terraform {
  required_version = ">= 1.3.7"

  backend "local" {}

  required_providers {
    nuon = {
      source = "nuonco/nuon"
      version = ">= 0.3.1"
    }
  }
}
