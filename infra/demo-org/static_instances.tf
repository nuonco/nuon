resource "nuon_build" "public_docker-build" {
  component_id = nuon_docker_build_component.public_docker.id
  git_ref = "main"
}

# statically manage what is running on an install, for your most valuable customers.
resource "nuon_deploy" "public_docker-deploy" {
  component_id = nuon_docker_build_component.public_docker.id
  build_id = nuon_build.public_docker-build.id
  install_id = nuon_install.demo-1.id
}
