resource "nuon_docker_build_component" "e2e" {
  name = "e2e_docker_build"
  app_id = nuon_app.main.id

  dockerfile = "Dockerfile"
  connected_repo = {
    directory = "components/go-httpbin"
    repo = "powertoolsdev/demo"
    branch = "main"
  }

  sync_only = true
}
