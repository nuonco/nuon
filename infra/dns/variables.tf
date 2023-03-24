variable "root_domain" {
  type        = string
  description = "root domain nuon.co"
  default     = "nuon.co"
}

variable "aws_region" {
  type    = string
  default = "us-west-1"
}

locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  name = "infra-nuon-dns"

  tags = {
    service   = local.name
    terraform = local.name
  }

  regions = {
    prod    = "us-west-2"
    stage   = "us-west-2"
    horizon = "us-west-2"
  }
}
