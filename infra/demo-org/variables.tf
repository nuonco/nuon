locals {
  name = terraform.workspace
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }
  region = "us-west-2"
  tags = {}
}
