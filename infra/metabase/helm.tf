locals {
  metabase = {
    value_file = "values/metabase.yaml"
    override_file = "values/${var.env}.yaml"
  }
}

#resource "helm_release" "metabase" {
  #namespace        = local.name
  #name             = "metabase"
  #create_namespace = true

  #repository = "./charts"
  #chart      = "metabase"
  #version    = "0.0.1"

  #values = [
    #file(local.metabase.value_file),
    #fileexists(local.metabase.override_file) ? file(local.metabase.override_file) : "",
  #]
#}
