locals {
  waypoint = {
    value_file = "values/waypoint.yaml"
  }
}

resource "helm_release" "waypoint" {
  namespace        = local.name
  name             = "waypoint"
  create_namespace = true

  repository = local.vars.chart.repo
  chart      = local.vars.chart.name
  version    = local.vars.chart.version

  values = [
    file(local.waypoint.value_file),

    // TODO(jm): add tags for environments
    yamlencode({
      waypoint = {
        tags        = ["env:${var.env}"]
        clusterName = var.env
      }
    })
  ]
}
