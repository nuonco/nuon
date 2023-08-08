resource "nuon_container_image_component" "e2e" {
  name = "e2e_external_image"
  app_id = nuon_app.main.id

  public = {
    image_url = "kennethreitz/httpbin"
    tag = "latest"
  }

  sync_only = true
}
