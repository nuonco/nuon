// nuon allows you to connect any helm chart in a connected or public repo to install in your application
resource "nuon_helm_chart_component" "demo-chart" {
  name = "Private Repo Helm Chart"
  app_id = nuon_app.main.id

  chart_name = "demo-chart"
  connected_repo = {
    directory = "charts/demo"
    repo = data.nuon_connected_repo.mono.name
    branch = data.nuon_connected_repo.mono.default_branch
  }

  // dynamically set env vars from another source
  dynamic "value" {
    for_each = local.default_helm_values
    iterator = ev
    content {
      name = ev.key
      value = ev.value
    }
  }
}

locals {
  default_helm_values = {
    "env.org_id" = "{{.nuon.org_id}}"
    "env.app_id" = "{{.nuon.app_id}}"
  }
}
