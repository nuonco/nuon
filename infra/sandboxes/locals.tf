locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  name                = "sandboxes"
  region              = "us-west-2"
  github_repository   = local.name
  github_organization = "powertoolsdev"

  tags = {
    environment = "shared"
    service     = local.name
    terraform   = local.name
  }
}
