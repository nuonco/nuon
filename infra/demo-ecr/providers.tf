locals {
  aws_settings = {
    region       = "us-west-2"
    account_name = "demo"
  }
}

provider "aws" {
  region = local.aws_settings.region

  default_tags {
    tags = {}
  }
}
