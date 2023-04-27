output "repository_arn" {
  value = module.ecr.repository_arn
}

output "registry_id" {
  value = module.ecr.repository_registry_id
}

output "repository_url" {
  value = module.ecr.repository_url
}

output "is_public" {
  value = false
}
