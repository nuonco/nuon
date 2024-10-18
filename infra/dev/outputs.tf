output "bucket_name" {
  description = "bucket name"
  value       = local.bucket_name
}

output "runner_dev" {
  description = "runner dev IAM role arn"
  value = module.runner_dev.iam_role_arn
}

output "dev_aws_ecr" {
  description = "dev ECR repo for container syncing testing"
  value = {
    nuon_access_role_arn = module.dev_ecr_access.iam_role_arn

    // repo access
    repo_name = module.dev_ecr.repository_name
    repo_arn = module.dev_ecr.repository_arn
    registry_id = module.dev_ecr.repository_registry_id
    image_url = module.dev_ecr.repository_url
  }
}

output "dev_install_access" {
  description = "dev access IAM roles"
  value = {
    "ecs" = module.ecs_access.iam_role_arn
    "eks" = module.eks_access.iam_role_arn
    "eks-byovc" = module.eks_byovpc_access.iam_role_arn
    "ecs-byovc" = module.ecs_byovpc_access.iam_role_arn
  }
}
