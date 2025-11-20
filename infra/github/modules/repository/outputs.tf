output "extra_ecr_repo_names" {
  description = "Full ECR repository names with prefix"
  value = [
    for repo in var.extra_ecr_repos : "${var.name}/${repo}"
  ]
}

output "lifecycle_policy" {
  description = "ECR lifecycle policy used by this module"
  value       = local.lifecycle_policy
}
