terraform {
  required_version = ">= 1.3.3"

  # NOTE: this module requires that a backend conf is used
  backend "s3" {}

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
  }
}
