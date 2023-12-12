locals {
  electric_sql_app_name = "electric_sql"
  electric_sql_app_id   = nuon_app.real[local.electric_sql_app_name].id

  auth_mode              = "insecure"
  logical_publisher_host = "electric"
  pg_proxy_password      = "electric_sql"

  db_user     = "electric_sql"
  db_password = "electric_sql"
  db_port     = 5432
  db_name     = local.electric_sql_app_name
}

resource "nuon_helm_chart_component" "helm_chart" {
  provider = nuon.real

  app_id     = local.electric_sql_app_id
  name       = "helm_chart"
  chart_name = "sync-service"

  connected_repo = {
    repo      = "powertoolsdev/mono"
    directory = "infra/customers-shared/electric-sql/components/helm-chart"
    branch    = "main"
  }

  value {
    name  = "DATABASE_URL"
    value = "postgresql://${local.db_user}:${local.db_password}@{{.nuon.components.rds_cluster.outputs.db_instance_endpoint}}:${local.db_port}/${local.db_name}"
  }

  value {
    name  = "AUTH_MODE"
    value = local.auth_mode
  }

  value {
    name  = "LOGICAL_PUBLISHER_HOST"
    value = local.logical_publisher_host
  }

  value {
    name  = "PG_PROXY_PASSWORD"
    value = local.pg_proxy_password
  }
}

resource "nuon_terraform_module_component" "rds_cluster" {
  provider = nuon.real

  app_id = local.electric_sql_app_id
  name   = "rds_cluster"

  connected_repo = {
    repo      = "powertoolsdev/mono"
    directory = "infra/customers-shared/electric-sql/components/rds-cluster"
    branch    = "main"
  }

  var {
    name  = "identifier"
    value = "electric-sql"
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
    name  = "vpc_security_group_id"
    value = "sg-02f07af3113dd063c"
  }

  var {
    name  = "subnet_id_one"
    value = "subnet-012a08391e6c093e1"
  }

  var {
    name  = "subnet_id_two"
    value = "subnet-0be8074c284cc6bdb"
  }
}

resource "nuon_install" "electric_sql_install" {
  provider = nuon.real

  app_id = local.electric_sql_app_id

  name         = "${local.electric_sql_app_name}-demo"
  region       = "us-east-1"
  iam_role_arn = "arn:aws:iam::949309607565:role/nuon-demo-org-prod-customer-iam-role"

  depends_on = [
    nuon_app_sandbox.real
  ]
}
