module "cert" {
  source = "../../../infra/modules/certificate"

  aws_region      = local.vars.region
  subdomain       = local.vars.subdomain
  use_root_domain = local.vars.use_root_domain
  env             = var.env
  service         = local.name
}

module "runner-cert" {
  source = "../../../infra/modules/certificate"

  aws_region      = local.vars.region
  subdomain       = local.vars.runner_subdomain
  use_root_domain = local.vars.use_root_domain
  env             = var.env
  service         = local.name
}
