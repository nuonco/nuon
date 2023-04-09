locals {
  datadog = {
    value_file = "values/datadog.yaml"
    # helm search repo datadog/datadog --versions to find latest versions
    version = "3.22.0"
  }
}

resource "helm_release" "datadog" {
  namespace        = local.name
  name             = "datadog-agent"
  create_namespace = true

  repository = "https://helm.datadoghq.com"
  chart      = "datadog"
  version    = local.datadog.version

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
