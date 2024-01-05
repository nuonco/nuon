resource "nuon_container_image_component" "e2e" {
  name   = "e2e_external_image"
  app_id = nuon_app.main.id

  dependencies = [
    nuon_terraform_module_component.e2e.id
  ]

  public = {
    image_url = "kennethreitz/httpbin"
    tag       = "latest"
  }
}
