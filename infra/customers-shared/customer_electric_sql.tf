locals {
  electric_sql_app_name = "electric_sql"
  electric_sql_app_id   = nuon_app.real[local.electric_sql_app_name].id

  db_user     = "electric_sql"
  db_password = "electric_sql"
  db_port     = 5432
  db_name     = local.electric_sql_app_name
}

resource "nuon_docker_build_component" "electric_sql" {
  provider = nuon.real

  app_id = local.electric_sql_app_id
  name   = "electric_sql"

  public_repo = {
    repo      = "https://github.com/electric-sql/electric"
    directory = "components/electric"
    branch    = "main"
  }

  depends_on = [
    nuon_terraform_module_component.rds_cluster,
  ]

  env_var {
    name  = "DATABASE_URL"
    value = "postgresql://${local.db_user}:${local.db_password}@{{.nuon.components.rds_cluster.outputs.db_instance_endpoint}}:${local.db_port}/${local.db_name}"
  }

  env_var {
    name  = "AUTH_MODE"
    value = "insecure"
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
