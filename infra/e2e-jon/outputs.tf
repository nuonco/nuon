output "ecs_access_iam_role" {
  value = module.ecs_access.iam_role_arn
}

output "eks_access_iam_role" {
  value = module.eks_access.iam_role_arn
}
