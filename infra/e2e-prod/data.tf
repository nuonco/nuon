data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

locals {
  region = "us-west-2"
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  tags = {
    terraform = "${local.name}"
  }
}
