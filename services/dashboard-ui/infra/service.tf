module "service" {
  source = "../../../infra/modules/service"

  name = local.name
  env  = var.env
  additional_iam_policies = []
}
