locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  region = "us-west-2"

  tags = {
    environment = "shared"
    terraform   = "infra-github"
  }
}

provider "github" {
  alias = "nuon"
  owner = "nuonco"

  app_auth {
    id              = var.powertools_app_id
    installation_id = var.powertools_app_installation_id
    pem_file        = var.powertools_app_pem_file
  }
}

provider "github" {
  owner = "powertoolsdev"

  app_auth {
    id              = var.powertools_app_id
    installation_id = var.powertools_app_installation_id
    pem_file        = var.powertools_app_pem_file
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

  assume_role {
    role_arn = "arn:aws:iam::${local.accounts.infra-shared-prod.id}:role/terraform"
  }

  default_tags {
    tags = local.tags
  }
}

data "aws_organizations_organization" "orgs" {
  provider = aws.mgmt
}

