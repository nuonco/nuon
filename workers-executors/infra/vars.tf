locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  name                   = "workers-executors"
  github_repository      = local.name
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

  vars = yamldecode(file("vars/${var.env}.yaml"))
}

variable "env" {
  type = string
}
