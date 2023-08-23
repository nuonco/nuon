variable "env" {
  type = string
}

variable "name" {
  type = string
}

variable "additional_iam_policies" {
  type = list(string)
  default = []
}

locals {
  accounts = {
    for acct in data.aws_organizations_organization.orgs.accounts : acct.name => { id : acct.id }
  }

  github_repository      = "mono"
  github_organization    = "powertoolsdev"
  terraform_organization = "launchpaddev"
  ecr_repository         = "${local.github_repository}/${var.name}"

  // helm configuration
  helm = {
    bucket_arn         = data.terraform_remote_state.chart_common.outputs.helm_bucket_arn
    bucket_url         = data.terraform_remote_state.chart_common.outputs.helm_bucket_url
    bucket_kms_key_arn = data.terraform_remote_state.chart_common.outputs.helm_bucket_kms_key_arn
  }

  tags = {
    environment = var.env
    service     = var.name
    terraform   = "${var.name}-${var.env}"
  }

  // variables that are environment specific
  vars = {
    prod = {
      region       = "us-west-2"
      cluster_name = "prod-nuon"
    }

    stage = {
      region       = "us-west-2"
      cluster_name = "stage-nuon"
    }
  }[var.env]
}
