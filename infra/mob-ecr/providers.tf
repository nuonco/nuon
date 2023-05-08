locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts :
    acct.name => acct.id
  }

  aws_settings = {
    region        = "us-west-2"
    demo_region = "us-east-1"
    account_name  = "demo"
  }
}

provider "aws" {
  region = local.aws_settings.demo_region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.demo}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

# this is the root account that the credentials have permissions for.
# use it to get list of accounts and pivot to the correct one
provider "aws" {
  alias  = "mgmt"
  region = local.aws_settings.region
}
