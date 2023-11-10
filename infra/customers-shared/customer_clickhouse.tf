resource "nuon_helm_chart_component" "clickhouse" {
  provider = nuon.sandbox

  name       = "clickhouse"
  app_id = nuon_app.sandbox["clickhouse"].id
  chart_name = "clickhouse"

  public_repo = {
    directory = "charts/clickhouse"
    repo      = "https://github.com/bitnami/charts"
    branch    = "main"
  }
}

resource "nuon_install" "clickhouse_install" {
  provider = nuon.sandbox

  app_id = nuon_app.sandbox["clickhouse"].id

  name         = "clickhouse-demo"
  region       = "us-east-1"
  iam_role_arn = "arn:aws:iam::949309607565:role/nuon-demo-install-access"
}
