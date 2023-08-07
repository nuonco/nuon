resource "nuon_terraform_module_component" "e2e" {
  name = "e2e-infra"
  app_id = nuon_app.main.id

  connected_repo = {
    directory = "services/e2e/infra"
    repo = "powertoolsdev/mono"
    branch = "main"
  }

  var {
    name = "install_id"
    value = "{{.nuon.install.id}}"
  }
  var {
    name = "region"
    value = "{{.nuon.install.sandbox.outputs.account.region}}"
  }
}
