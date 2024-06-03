resource "nuon_app_input" "main" {
  app_id = nuon_app.main.id

  dynamic "group" {
    for_each = var.groups

    content {
      name = group.value.name
      description = group.value.description
      display_name = group.value.display_name
    }
  }

  dynamic "input" {
    for_each = var.inputs

    content {
      name         = input.value.name
      description  = input.value.description
      default      = input.value.default
      required     = input.value.required
      display_name = input.value.display_name
      sensitive    = input.value.sensitive
    }
  }
}
