locals {
  name = "ctl-api"
}

module "service" {
  source = "../../../infra/modules/service"

  name      = local.name
  env       = var.env
  namespace = local.namespace
  additional_iam_policies = [
    aws_iam_policy.additional_permissions.arn
  ]
}
