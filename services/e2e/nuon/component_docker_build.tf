resource "nuon_docker_build_component" "e2e" {
  count = var.create_components ? 1 : 0

  name   = "${var.component_prefix}e2e_docker_build"
  app_id = nuon_app.main.id

  dockerfile = "Dockerfile"
  dependencies = [
    nuon_container_image_component.e2e[0].id,
  ]
  connected_repo = {
    directory = "components/go-httpbin"
    repo      = "powertoolsdev/demo"
    branch    = "main"
  }
}
