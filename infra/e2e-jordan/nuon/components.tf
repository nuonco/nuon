locals {
  electric_sql_app_name            = "electric_sql"
  electric_sql_app_name_hyphenated = "electric-sql"

  auth_mode              = "insecure"
  logical_publisher_host = local.electric_sql_app_name
  pg_proxy_password      = local.electric_sql_app_name

  db_user     = local.electric_sql_app_name
  db_password = local.electric_sql_app_name
  db_port     = 5432
  db_name     = local.electric_sql_app_name
  db_url      = "postgresql://${local.db_user}:${local.db_password}@{{.nuon.components.rds_cluster.outputs.db_instance_endpoint}}/${local.db_name}"
}

resource "nuon_terraform_module_component" "aws_ecs_sync_service" {
  name   = "sync_service"
  app_id = nuon_app.main.id

  connected_repo = {
    directory = "infra/e2e-jordan/components/sync-service"
    repo      = "powertoolsdev/mono"
    branch    = "main"
  }


  # service config

  var {
    name  = "database_url"
    value = local.db_url
  }

  var {
    name  = "auth_mode"
    value = local.auth_mode
  }

  var {
    name  = "pg_proxy_password"
    value = local.pg_proxy_password
  }


  # hosting config

  var {
    name  = "vpc_id"
    value = "{{.nuon.install.sandbox.outputs.vpc.id}}"
  }

  var {
    name  = "subnet_ids"
    value = "{{.nuon.install.sandbox.outputs.vpc.private_subnet_ids}}"
  }

  var {
    name  = "cluster_arn"
    value = "{{.nuon.install.sandbox.outputs.ecs_cluster.arn}}"
  }


  # networking config

  var {
    name  = "domain_name"
    value = "electric.{{.nuon.install.sandbox.outputs.public_domain.name}}"
  }

  var {
    name  = "zone_id"
    value = "{{.nuon.install.sandbox.outputs.public_domain.zone_id}}"
  }
}

resource "nuon_terraform_module_component" "aws_ecs_rds_cluster" {
  app_id = nuon_app.main.id

  name = "rds_cluster"

  connected_repo = {
    directory = "infra/e2e-jordan/components/rds-cluster"
    repo      = "powertoolsdev/mono"
    branch    = "main"
  }

  var {
    name  = "identifier"
    value = nuon_app.main.id
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
