locals {
  electric_sql_app_name = "electric_sql"
  electric_sql_app_id   = nuon_app.sandbox[local.electric_sql_app_name].id

  auth_mode              = "insecure"
  logical_publisher_host = "electric"
  pg_proxy_password      = "electric_sql"

  db_user     = "electric_sql"
  db_password = "electric_sql"
  db_port     = 5432
  db_name     = local.electric_sql_app_name
}

resource "nuon_terraform_module_component" "certificate" {
  provider = nuon.sandbox

  app_id = local.electric_sql_app_id
  name   = "certificate"

  connected_repo = {
    repo      = "powertoolsdev/mono"
    directory = "infra/customers-shared/electric-sql/components/certificate"
    branch    = "main"
  }

  var {
    name  = "domain_name"
    value = "nlb.{{.nuon.install.sandbox.outputs.public_domain.name}}"
  }

  var {
    name  = "zone_id"
    value = "{{.nuon.install.sandbox.outputs.public_domain.zone_id}}"
  }
}

resource "nuon_helm_chart_component" "helm_chart" {
  provider = nuon.sandbox

  app_id     = local.electric_sql_app_id
  name       = "helm_chart"
  chart_name = "sync-service"

  connected_repo = {
    repo      = "powertoolsdev/mono"
    directory = "infra/customers-shared/electric-sql/components/helm-chart"
    branch    = "main"
  }

  value {
    name  = "env.DATABASE_URL"
    value = "postgresql://${local.db_user}:${local.db_password}@{{.nuon.components.rds_cluster.outputs.db_instance_endpoint}}/${local.db_name}"
  }

  value {
    name  = "env.AUTH_MODE"
    value = local.auth_mode
  }

  value {
    name  = "env.LOGICAL_PUBLISHER_HOST"
    value = local.logical_publisher_host
  }

  value {
    name  = "env.PG_PROXY_PASSWORD"
    value = local.pg_proxy_password
  }

  value {
    name  = "api.ingresses.public_domain"
    value = "api.{{.nuon.install.public_domain}}"
  }

  value {
    name  = "api.ingresses.internal_domain"
    value = "api.{{.nuon.install.internal_domain}}"
  }

  value {
    name  = "api.nlbs.public_domain"
    value = "nlb.{{.nuon.install.public_domain}}"
  }

  value {
    name  = "api.nlbs.internal_domain"
    value = "nlb.internal.{{.nuon.install.internal_domain}}"
  }

  value {
    name  = "api.nlbs.public_domain_certificate_arn"
    value = "{{.nuon.components.certificate.outputs.public_domain_certificate_arn}}"
  }
}

resource "nuon_terraform_module_component" "rds_cluster" {
  provider = nuon.sandbox

  app_id = local.electric_sql_app_id
  name   = "rds_cluster"

  connected_repo = {
    repo      = "powertoolsdev/mono"
    directory = "infra/customers-shared/electric-sql/components/rds-cluster"
    branch    = "main"
  }

  var {
    name  = "identifier"
    value = "skynet-inc-db"
  }

  var {
    name  = "db_name"
    value = local.db_name
  }

  var {
    name  = "username"
    value = local.db_user
  }

  var {
    name  = "password"
    value = local.db_password
  }

  var {
    name  = "port"
    value = local.db_port
  }

  var {
    name  = "subnet_id_one"
    value = "{{index .nuon.install.sandbox.outputs.vpc.private_subnet_ids 0}}"
  }

  var {
    name  = "subnet_id_two"
    value = "{{index .nuon.install.sandbox.outputs.vpc.private_subnet_ids 1}}"
  }

  var {
    name  = "vpc_id"
    value = "{{.nuon.install.sandbox.outputs.vpc.id}}"
  }

  var {
    name  = "zone_id"
    value = "{{.nuon.install.sandbox.outputs.internal_domain.zone_id}}"
  }
}

resource "nuon_install" "electric_sql_us_east_2" {
  provider = nuon.sandbox

  app_id = local.electric_sql_app_id

  name         = "${local.electric_sql_app_name}_us_east_2"
  region       = "us-east-2"
  iam_role_arn = "arn:aws:iam::949309607565:role/nuon-demo-org-prod-customer-iam-role"

  depends_on = [
    nuon_app_sandbox.sandbox
  ]
}
