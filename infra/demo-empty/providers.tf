locals {
  aws_settings = {
    region       = "us-west-2"
  }
}

provider "aws" {
  region = local.aws_settings.region
  default_tags {
    tags = {}
  }
}
