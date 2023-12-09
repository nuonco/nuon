locals {
  electric_sql_app_name    = "electric_sql"
  electric_sql_app_id      = nuon_app.sandbox[local.electric_sql_app_name].id
  electric_sql_repo_url    = "https://github.com/nuonco/customer-electric-sql"
  electric_sql_repo_branch = "main"
}

resource "nuon_docker_build_component" "electric_sql" {
  provider = nuon.sandbox
  app_id   = local.electric_sql_app_id

  name = local.electric_sql_app_name

  connected_repo = {
    repo      = local.electric_sql_repo_url
    branch    = local.electric_sql_repo_branch
    directory = "components/electric_sql"
  }
}

resource "nuon_terraform_module_component" "electric_sql_rds" {
  provider = nuon.sandbox
  app_id   = local.electric_sql_app_id

  name = "electric_sql_rds"

  connected_repo = {
    repo      = local.electric_sql_repo_url
    branch    = local.electric_sql_repo_branch
    directory = "components/rds"
  }
}

resource "nuon_install" "electric_sql_install" {
  provider = nuon.sandbox

  app_id = nuon_app.sandbox["electric_sql"].id

  name         = "electric_sql"
  region       = "us-west-2"
  iam_role_arn = "arn:aws:iam::949309607565:role/nuon-demo-install-access"
}

