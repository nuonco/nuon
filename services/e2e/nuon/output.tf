output "app_id" {
  value = nuon_app.main.id
}

output "app" {
  value = nuon_app.main
}

output "app_installer_slug" {
  value = nuon_app_installer.main.slug
}

output "app_installer" {
  value = nuon_app_installer.main
}

output "component_ids" {
  value = [
    nuon_docker_build_component.e2e.id,
    nuon_container_image_component.e2e.id,
    nuon_helm_chart_component.e2e.id,
    nuon_terraform_module_component.e2e.id,
    nuon_job_component.e2e.id,
  ]
}

output "components" {
  value = {
    "docker_build" : nuon_docker_build_component.e2e,
    "container_image" : nuon_container_image_component.e2e,
    "helm_chart" : nuon_helm_chart_component.e2e,
    "terraform_module" : nuon_terraform_module_component.e2e.id,
    "job" : nuon_job_component.e2e.id,
  }
}

output "install_ids" {
  value = concat(
    nuon_install.east_2.*.id,
    nuon_install.east_1.*.id,
    nuon_install.west_2.*.id,
  )
}

output "installs" {
  value = {
    "east-1" : nuon_install.east_1,
    "west-2" : nuon_install.west_2,
    "east-2" : nuon_install.east_2,
  }
}
