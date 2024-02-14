resource "nuon_app_sandbox" "main" {
  app_id            = nuon_app.main.id
  terraform_version = "v1.6.3"

  public_repo = {
    repo      = var.sandbox_repo
    branch    = var.sandbox_branch
    directory = var.sandbox_dir
  }

  dynamic "var" {
    for_each = var.inputs
    content {
      name  = var.value.name
      value = var.value.interpolation
    }
  }
}
