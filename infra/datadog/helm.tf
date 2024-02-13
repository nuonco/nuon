locals {
  datadog = {
    value_file = "values/datadog.yaml"
    override_file = "values/${var.env}.yaml"
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
    fileexists(local.datadog.override_file) ? file(local.datadog.override_file) : "",

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
