terraform {
  required_version = ">= 1.7.5"

  backend "s3" {}

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.62.0"
    }
  }
}
