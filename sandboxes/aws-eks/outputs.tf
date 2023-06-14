output "cluster_arn" {
  description = "The Amazon Resource Name (ARN) of the cluster"
  value       = module.eks.cluster_arn
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = module.eks.cluster_certificate_authority_data
}

output "cluster_endpoint" {
  description = "Endpoint for your Kubernetes API server"
  value       = module.eks.cluster_endpoint
}

output "cluster_name" {
  description = "The name of the EKS cluster. Will block on cluster creation until the cluster is really ready"
  value       = module.eks.cluster_name
}

output "cluster_platform_version" {
  description = "Platform version for the cluster"
  value       = module.eks.cluster_platform_version
}

output "cluster_status" {
  description = "Status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`"
  value       = module.eks.cluster_status
}

output "ecr_registry_id" {
  description = "The ECR registry information"
  value       = module.ecr.repository_registry_id
}

output "ecr_registry_arn" {
  description = "The ECR registry information"
  value       = module.ecr.repository_arn
}

output "ecr_registry_url" {
  description = "The ECR registry information"
  value       = module.ecr.repository_url
}

output "odr_iam_role_arn" {
  description = "iam role arn of the odr's IAM role which grants permissions to ECR"
  value       = module.odr_iam_role.iam_role_arn
}
