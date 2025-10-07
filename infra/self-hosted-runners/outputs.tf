output "controller_release_name" {
  description = "The name of the controller Helm release"
  value       = helm_release.gha_runner_controller.name
}

output "controller_namespace" {
  description = "The namespace where the controller is deployed"
  value       = helm_release.gha_runner_controller.namespace
}

output "scale_set_names" {
  description = "The names of all deployed scale sets"
  value       = keys(helm_release.gha_runner_scale_sets)
}

output "scale_set_releases" {
  description = "Information about all scale set releases"
  value = {
    for k, v in helm_release.gha_runner_scale_sets : k => {
      name      = v.name
      namespace = v.namespace
      version   = v.version
    }
  }
}