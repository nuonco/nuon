provider "aws" {
  region = "us-east-1"

  assume_role {
    role_arn = local.role_arn
  }

  default_tags {
    tags = local.tags
  }
}
