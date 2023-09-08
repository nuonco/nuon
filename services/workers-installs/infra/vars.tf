locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  name                   = "workers-installs"
  github_repository      = "mono"
  github_organization    = "powertoolsdev"
  terraform_organization = "launchpaddev"

  tags = {
    environment = var.env
    service     = local.name
    terraform   = "${local.name}-${var.env}"
  }

  vars = yamldecode(file("vars/${var.env}.yaml"))
}

variable "env" {
  type = string
}
