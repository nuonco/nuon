resource "nuon_job_component" "e2e" {
  count = var.create_components ? 1 : 0

  name   = "${var.component_prefix}e2e_job"
  app_id = nuon_app.main.id
  dependencies = [
    nuon_docker_build_component.e2e[0].id,
    nuon_container_image_component.e2e[0].id,
  ]

  image_url = "{{.nuon.components.e2e_docker_build.image.repository.uri}}"
  tag       = "{{.nuon.components.e2e_docker_build.image.tag}}"
  cmd       = ["printenv"]
  args      = [""]

  env_var {
    name  = "NUON_APP_ID"
    value = "{{.nuon.app.id}}"
  }

  env_var {
    name  = "NUON_ORG_ID"
    value = "{{.nuon.org.id}}"
  }

  env_var {
    name  = "NUON_INSTALL_ID"
    value = "{{.nuon.install.id}}"
  }
}
