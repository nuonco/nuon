locals {
  electric_sql_app_name = "electric_sql"
  electric_sql_app_id   = nuon_app.real[local.electric_sql_app_name].id
}

resource "nuon_docker_build_component" "electric_sql" {
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
}

resource "nuon_terraform_module_component" "rds_cluster" {
  app_id = local.electric_sql_app_id
  name   = "rds_cluster"

  public_repo = {
    repo      = "https://github.com/nuonco/customer-electric-sql"
    directory = "components/rds_cluster"
    branch    = "main"
  }

  var {
    name  = "identifier"
    value = "electric_sql"
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
    value = "electric_sql"
  }

  var {
    name  = "username"
    value = "electric_sql"
  }

  var {
    name  = "port"
    value = "3306"
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
