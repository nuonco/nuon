module "service" {
  source = "../../../infra/modules/service"

  name                    = local.name
  namespace               = local.namespace
  env                     = var.env
  additional_iam_policies = []
}
