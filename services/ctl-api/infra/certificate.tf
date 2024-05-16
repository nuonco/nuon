// NOTE(jm): this is a replacement for the legacy cert, and uses the alias name
module "cert" {
  source = "../../../infra/modules/certificate"

  aws_region      = local.vars.region
  subdomain       = local.vars.subdomain
  use_root_domain = local.vars.use_root_domain
  env             = var.env
  service         = local.name
}

