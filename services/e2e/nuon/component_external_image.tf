resource "nuon_container_image_component" "e2e" {
  count = var.create_components ? 1 : 0

  name   = "${var.component_prefix}e2e_external_image"
  app_id = nuon_app.main.id

  dependencies = [
    nuon_terraform_module_component.e2e[0].id
  ]

  public = {
    image_url = "kennethreitz/httpbin"
    tag       = "latest"
  }
}
