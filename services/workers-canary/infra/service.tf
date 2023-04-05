module "service" {
  source = "../../../infra/modules/service"

  name = "workers-canary"
  env = var.env
}
