locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts :
    acct.name => acct.id
  }

  tf_role_arn = "arn:aws:iam::${local.accounts[var.env]}:role/terraform"
  account_name = "public"
}

# this is the root account that the credentials have permissions for.
# use it to get list of accounts and pivot to the correct one
provider "aws" {
  region = local.vars.region
  alias  = "mgmt"
}

provider "aws" {
  region = local.vars.region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[local.account_name]}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}
