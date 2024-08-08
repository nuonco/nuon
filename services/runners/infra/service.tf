module "service" {
  source = "../../../infra/modules/service"

  name = "runners"
  env  = var.env
  additional_iam_policies = [
    aws_iam_policy.additional_permissions.arn
  ]
}
