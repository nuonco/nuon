locals {
  default_tf_vars = {
    repo_name = "test-demo-tf"
  }
}

// nuon allows you to connect any terraform module in a connected or public repo.
// The terraform _actually_ runs inside of your customer's cloud account, so you can do things like manage internal
// resources etc.
resource "nuon_terraform_module_component" "demo-ecr" {
  name = "terraform_infra"
  app_id = nuon_app.main.id

  connected_repo = {
    directory = "infra/demo-ecr"
    repo = "powertoolsdev/mono"
    branch = "main"
  }

  // dynamically set env vars from another source
  dynamic "var" {
    for_each = local.default_tf_vars
    iterator = ev
    content {
      name = ev.key
      value = ev.value
    }
  }

  var {
    name = "ecr_repo_name"
    value = "{{.nuon.app_id}}"
  }
}
