output "gh_role_arn" {
  value = module.service.gh_role_arn
}

output "region" {
  value = module.service.region
}

output "eks_role_arn" {
  value = module.service.eks_role_arn
}

output "ecr_repository_url" {
  value = module.service.ecr_repository_url
}

output "ecr_registry_id" {
  value = module.service.ecr_registry_id
}

output "cluster_name" {
  value = module.service.cluster_name
}

output "cluster_gh_role_arn" {
  value = module.service.cluster_gh_role_arn
}

