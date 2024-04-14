locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  bucket_name = "nuon-dev"
  name        = "dev"
  region      = "us-west-2"

  tags = {
    service     = local.name
    terraform   = "${local.name}"
  }
}
