locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  tags = {
    environment = var.env
    service     = var.service
    terraform   = "${var.service}-${var.env}"
  }
}

