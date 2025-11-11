locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts :
    acct.name => acct.id
  }

  aws_settings = {
    region       = "us-west-2"
    account_name = "infra-shared-prod"
  }
}

# this is the root account that the credentials have permissions for.
# use it to get list of accounts and pivot to the correct one
provider "aws" {
  region = local.aws_settings.region
  alias  = "mgmt"
}

provider "aws" {
  region = local.aws_settings.region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[local.aws_settings.account_name]}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}
