locals {
  honeyhive_app_name = "honeyhive"
  honeyhive_app_id   = nuon_app.sandbox[local.honeyhive_app_name].id
}

resource "nuon_terraform_module_component" "document_db" {
  provider = nuon.sandbox

  name   = "document_db"
  app_id = local.honeyhive_app_id

  connected_repo = {
    repo      = "powertoolsdev/mono"
    directory = "infra/customers-shared/honeyhive/components/document-db"
    branch    = "main"
  }

  var {
    name  = "vpc_id"
    value = "{{.nuon.install.sandbox.outputs.vpc.id}}"
  }

  var {
    name  = "namespace"
    value = local.honeyhive_app_name
  }

  var {
    name  = "stage"
    value = "production"
  }

  var {
    name  = "name"
    value = local.honeyhive_app_name
  }

  var {
    name  = "cluster_size"
    value = 1
  }

  var {
    name  = "master_username"
    value = local.honeyhive_app_name
  }

  var {
    name  = "master_password"
    value = local.honeyhive_app_name
  }

  var {
    name  = "instance_class"
    value = "db.r6g.large"
  }

  var {
    name  = "subnet_one"
    value = "{{index .nuon.install.sandbox.outputs.vpc.private_subnet_ids 0}}"
  }

  var {
    name  = "subnet_two"
    value = "{{index .nuon.install.sandbox.outputs.vpc.private_subnet_ids 1}}"
  }

  var {
    name  = "zone_id"
    value = "{{.nuon.install.sandbox.outputs.public_domain.zone_id}}"
  }
}
