locals {
  buildkit = {
    value_file    = "values/buildkit.yaml"
    override_file = "values/${var.env}.yaml"
  }
}

resource "helm_release" "buildkit" {
  namespace = local.vars.buildkit_namespace
  name      = "buildkit"

  create_namespace = false

  repository = "./charts"
  chart      = "buildkit"
  version    = "0.0.1"

  set {
    name  = "nodePool.name"
    value = local.vars.node_pool_name
  }

  values = [
    file(local.buildkit.value_file),
    fileexists(local.buildkit.override_file) ? file(local.buildkit.override_file) : "",
  ]
}
