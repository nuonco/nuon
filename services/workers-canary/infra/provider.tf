data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

locals {
  region = "us-west-2"
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  tags = {
    service   = local.name
    terraform = "${local.name}-${var.env}"
  }
}

provider "aws" {
  region = local.region
  alias  = "mgmt"
  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region
  alias  = "canary"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.canary.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region
  alias  = "infra-shared-prod"

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.infra-shared-prod.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

provider "aws" {
  region = local.region

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts[var.env].id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}
