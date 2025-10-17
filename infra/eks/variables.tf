locals {
  # TODO(jdt): rename this
  workspace_trimmed = "${var.account}-${var.pool}"
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts :
    acct.name => acct
  }
  target_account_id = local.accounts[var.account].id

  terraform_organization = "nuonco"

  tags = {
    environment = var.account
    pool        = var.pool
    tier        = local.vars.tier
    terraform   = "infra-eks-${var.account}-${var.pool}"
  }

  github_repository   = "mono"
  github_organization = "powertoolsdev"

  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
    file("vars/${local.workspace_trimmed}.yaml"),
  ]
}

variable "account" {
  description = "The AWS account to launch the cluster into"
  type        = string
}

variable "pool" {
  description = "The cluster pool"
  type        = string
}

variable "vantage_api_token" {
  description = "The vantage api token."
  type        = string
  sensitive   = true
}
