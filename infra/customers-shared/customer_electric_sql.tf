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

resource "nuon_container_image_component" "sync_service" {
  provider = nuon.real

  app_id = local.electric_sql_app_id
  name   = "sync_service"

  public = {
    image_url = "electricsql/electric"
    tag       = "latest"
  }

  env_var {
    name  = "DATABASE_URL"
    value = "postgresql://${local.db_user}:${local.db_password}@{{.nuon.components.rds_cluster.outputs.db_instance_endpoint}}:${local.db_port}/${local.db_name}"
  }

  env_var {
    name  = "AUTH_MODE"
    value = local.auth_mode
  }

  env_var {
    name  = "LOGICAL_PUBLISHER_HOST"
    value = local.logical_publisher_host
  }

  env_var {
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
    value = local.electric_sql_app_name
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
    name  = "subnet_id"
    value = "subnet-012a08391e6c093e1"
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
