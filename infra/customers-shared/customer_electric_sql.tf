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

resource "nuon_container_image_component" "electric_sql" {
  provider = nuon.real

  app_id = local.electric_sql_app_id
  name   = "electric_sql"

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

  public_repo = {
    repo      = "https://github.com/nuonco/customer-electric-sql"
    directory = "components/rds_cluster"
    branch    = "main"
  }

  var {
    name  = "identifier"
    value = local.electric_sql_app_name
  }

  var {
    name  = "engine_version"
    value = "5.7"
  }

  var {
    name  = "instance_class"
    value = "db.t3a.large"
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
    name  = "iam_database_authentication_enabled"
    value = true
  }

  var {
    name  = "vpc_security_group_ids"
    value = "[{{.nuon.install.sandbox.outputs.eks.node_security_group_id}}]"
  }

  var {
    name  = "subnet_ids"
    value = "{{.nuon.install.sandbox.outputs.vpc.private_subnet_ids}}"
  }
}
