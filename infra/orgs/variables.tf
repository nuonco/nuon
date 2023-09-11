locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  additional_install_role_eks_principals = [
    "eks-workers-installs",
    "eks-workers-orgs",
  ]

  org_id = data.aws_organizations_organization.orgs.id

  name                   = "infra-installations"
  region                 = "us-west-2"
  target_account         = "infra-shared-${var.env}"
  terraform_organization = "nuonco"
  org_account_id         = local.accounts["orgs-${var.env}"].id

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
