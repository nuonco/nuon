# this demo shows off some static instances
resource "nuon_build" "public_docker-build" {
  component_id = nuon_docker_build_component.public_docker.id
  git_ref = "main"
}

resource "nuon_deploy" "public_docker-deploy" {
  component_id = nuon_docker_build_component.public_docker.id
  build_id = nuon_docker_build_component.public_docker.id
  install_id = nuon_install.demo-stage.id
}
