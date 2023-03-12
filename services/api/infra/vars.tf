data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
    file("vars/${var.env}.yaml"),
  ]
}

locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  target_account_id = local.accounts[var.env].id

  name                   = "api"
  pool                   = "nuon"
  github_repository      = "mono"
  github_organization    = "powertoolsdev"
  terraform_organization = "launchpaddev"

  helm_bucket_arn         = data.terraform_remote_state.chart_common.outputs.helm_bucket_arn
  helm_bucket_url         = data.terraform_remote_state.chart_common.outputs.helm_bucket_url
  helm_bucket_kms_key_arn = data.terraform_remote_state.chart_common.outputs.helm_bucket_kms_key_arn

  tags = {
    environment = var.env
    service     = local.name
    terraform   = "${local.name}-${var.env}"
  }

  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

variable "env" {
  type = string
}
