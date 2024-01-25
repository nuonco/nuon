data "nuon_builtin_sandbox" "main" {
  name = "aws-eks"
}

resource "nuon_app_sandbox" "main" {
  app_id            = nuon_app.main.id
  terraform_version = "v1.6.3"

  public_repo = {
    repo      = var.sandbox_repo
    branch    = var.sandbox_branch
    directory = var.sandbox_dir
  }

  dynamic "var" {
    for_each = var.install_inputs
    content {
      name  = var.value.name
      value = var.value.interpolation
    }
  }
}
