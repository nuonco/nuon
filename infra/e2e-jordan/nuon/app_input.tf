resource "nuon_app_input" "main" {
  app_id = nuon_app.main.id

  dynamic "input" {
    for_each = var.install_inputs
    content {
      name        = input.value.name
      description = input.value.description
      default     = input.value.default
      required    = input.value.required
      display_name = input.value.name
    }
  }
}
