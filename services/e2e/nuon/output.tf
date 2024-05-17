output "app_id" {
  value = nuon_app.main.id
}

output "app" {
  value = nuon_app.main
}

output "installer" {
  value = nuon_installer.main
}

output "component_ids" {
  value = !var.create_components ? [] : [
    nuon_docker_build_component.e2e[0].id,
    nuon_container_image_component.e2e[0].id,
    nuon_helm_chart_component.e2e[0].id,
    nuon_terraform_module_component.e2e[0].id,
    nuon_job_component.e2e[0].id,
  ]
}

output "components" {
  value = !var.create_components ? {} : {
    "docker_build" : nuon_docker_build_component.e2e[0].id,
    "container_image" : nuon_container_image_component.e2e[0].id,
    "helm_chart" : nuon_helm_chart_component.e2e[0].id,
    "terraform_module" : nuon_terraform_module_component.e2e[0].id,
    "job" : nuon_job_component.e2e[0].id,
  }
}

output "install_ids" {
  value = nuon_install.main[*].id
}

output "installs" {
  value = nuon_install.main[*].name
}
