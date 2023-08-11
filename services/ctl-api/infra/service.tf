module "service" {
  source = "../../../infra/modules/service"

  name = "ctl-api"
  env  = var.env
}
