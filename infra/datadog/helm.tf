locals {
  datadog = {
    value_file = "values/datadog.yaml"
  }
}

resource "helm_release" "datadog" {
  namespace        = local.name
  name             = "datadog-agent"
  create_namespace = true

  repository = local.vars.chart.repo
  chart      = local.vars.chart.name
  version    = local.vars.chart.version

  values = [
    file(local.datadog.value_file),

    // TODO(jm): add tags for environments
    yamlencode({
      datadog = {
        apiKey      = var.datadog_api_key
        tags        = ["env:${var.env}"]
        clusterName = var.env
      }
    })
  ]
}
