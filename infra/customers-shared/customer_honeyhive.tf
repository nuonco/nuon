locals {
  honeyhive_app_name = "honeyhive"
  honeyhive_app_id   = nuon_app.sandbox[local.honeyhive_app_name].id
}

resource "nuon_terraform_module_component" "document_db" {
  provider = nuon.sandbox

  name   = "document_db"
  app_id = local.honeyhive_app_id

  public_repo = {
    repo      = "powertoolsdev/mono"
    directory = "infra/customers-shared/honeyhive/components/document-db"
    branch    = "main"
  }

  var {
    name  = "vpc_id"
    value = "{{.nuon.install.inputs.vpc_id}}"
  }

  var {
    name  = "namespace"
    value = "honeyhive"
  }

  var {
    name  = "stage"
    value = "production"
  }

  var {
    name  = "name"
    value = "{{.nuon.install.inputs.install_name}}"
  }

  var {
    name  = "cluster_size"
    value = 3
  }

  var {
    name  = "master_username"
    value = "honeyhive"
  }

  var {
    name  = "master_password"
    value = "password"
  }

  var {
    name  = "instance_class"
    value = "db.r4.large"
  }

  var {
    name  = "subnet_ids"
    value = "{{.nuon.install.sandbox.outputs.vpc.private_subnet_ids}}"
  }

  var {
    name  = "allowed_security_groups"
    value = "default"
  }

  var {
    name  = "zone_id"
    value = "{{.nuon.install.sandbox.outputs.public_domain.zone_id}}"
  }
}
