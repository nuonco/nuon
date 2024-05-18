data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
    file("vars/${var.env}.yaml"),
  ]
}

locals {
  name                   = "wiki"
  vars                   = yamldecode(data.utils_deep_merge_yaml.vars.output)
  terraform_organization = "nuonco"
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }
  region = "us-west-2"
  tags = {
    service   = local.name
    terraform = "${local.name}-${var.env}"
    env       = var.env
  }
}

variable "env" {
  type        = string
  description = "env"
}

variable "tfe_token" {
  type        = string
  description = "tfe_token"
}
