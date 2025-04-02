locals {
  metabase = {
    value_file = "values/metabase.yaml"
    override_file = "values/${var.env}.yaml"
  }
}

resource "helm_release" "metabase" {
  namespace        = "metabase"
  name             = "metabase"

  create_namespace = true

  repository = "./charts"
  chart      = "metabase"
  version    = "0.0.1"

  values = [
    file(local.metabase.value_file),
    fileexists(local.metabase.override_file) ? file(local.metabase.override_file) : "",
    yamlencode(
      {
        env = {
          "MB_DB_HOST" = module.primary.db_instance_address,
        },
        },
      ),
  ]
}
