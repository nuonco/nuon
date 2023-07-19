locals {
  waypoint = {
    value_file = "values/waypoint.yaml"
  }
}

resource "helm_release" "waypoint" {
  namespace        = local.name
  name             = "waypoint"
  create_namespace = true

  repository = local.vars.chart.path
  chart      = local.vars.chart.name
  version    = local.vars.chart.version

  values = [
    file(local.waypoint.value_file),
    yamlencode({
      server = {
        domain = "waypoint.${var.env}.${local.vars.public_root_domain}"
      }
    })
  ]
}
