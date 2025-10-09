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

output "github_secret_name" {
  description = "The name of the GitHub token secret"
  value       = local.vars.github_secret_name
}

output "github_secret_namespace" {
  description = "The namespace where the GitHub token secret is deployed"
  value       = local.vars.runner_namespace
}

output "runner_iam_roles" {
  description = "IAM roles created for each runner scale set"
  value = {
    for k, v in aws_iam_role.runner_scale_set_roles : k => {
      name = v.name
      arn  = v.arn
    }
  }
}

output "runner_service_accounts" {
  description = "Service accounts created for each runner scale set"
  value = {
    for k, v in aws_iam_role.runner_scale_set_roles : k => {
      name      = "${k}-runner"
      namespace = local.vars.runner_namespace
      role_arn  = v.arn
    }
  }
}