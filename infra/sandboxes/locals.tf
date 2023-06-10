locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  name                = "sandboxes"
  repo                = "mono"
  region              = "us-west-2"
  github_repository   = local.repo
  github_organization = "powertoolsdev"

  tags = {
    environment = "shared"
    service     = local.repo
    terraform   = local.repo
  }
}
