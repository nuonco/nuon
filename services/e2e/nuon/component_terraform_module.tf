resource "nuon_terraform_module_component" "e2e" {
  count = var.create_components ? 1 : 0

  name              = "${var.component_prefix}e2e_infra"
  app_id            = nuon_app.main.id
  var_name = "e2e_terraform"
  terraform_version = "1.6.3"

  dependencies = []

  connected_repo = {
    directory = "services/e2e/infra"
    repo      = "powertoolsdev/mono"
    branch    = "main"
  }

  env_var {
    name  = "install_id"
    value = "{{.nuon.install.id}}"
  }

  env_var {
    name  = "AWS_REGION"
    value = "{{.nuon.install.sandbox.outputs.account.region}}"
  }

  var {
    name  = "install_id"
    value = "{{.nuon.install.id}}"
  }
  var {
    name  = "region"
    value = "{{.nuon.install.sandbox.outputs.account.region}}"
  }

  var {
    name  = "public_domain"
    value = "{{.nuon.install.public_domain}}"
  }

  var {
    name  = "public_domain_zone_id"
    value = "{{.nuon.install.sandbox.outputs.public_domain.zone_id}}"
  }
}
