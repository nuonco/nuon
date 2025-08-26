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

module "internal-cert" {
  source = "../../../infra/modules/internal-certificate"

  aws_region = local.vars.region
  subdomain  = "ctl"
  domain     = local.vars.internal_root_domain
  env        = var.env
  service    = local.name
}
