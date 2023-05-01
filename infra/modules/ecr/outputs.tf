output "repository_arn" {
  value = module.ecr.repository_arn
}

output "registry_id" {
  value = module.ecr.repository_registry_id
}

output "repository_url" {
  value = module.ecr.repository_url
}

output "registry_url" {
  value = "${module.ecr.repository_registry_id}.dkr.ecr.${var.region}.amazonaws.com"
}

output "region" {
  value = var.region
}

output "is_public" {
  value = false
}
