locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  bucket_name = "nuon-dev"
  name        = "dev"
  env         = "dev"
  region      = "us-west-2"

  tags = {
    environment = local.env
    service     = local.name
    terraform   = "${local.name}"
  }
}
