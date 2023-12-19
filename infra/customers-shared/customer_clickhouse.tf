resource "nuon_helm_chart_component" "clickhouse" {
  provider = nuon.sandbox

  name       = "clickhouse"
  app_id     = nuon_app.sandbox["clickhouse"].id
  chart_name = "clickhouse"

  public_repo = {
    directory = "charts/clickhouse"
    repo      = "https://github.com/bitnami/charts"
    branch    = "main"
  }
}
