output "gh_role_arn" {
  value = module.github_actions.iam_role_arn
}

output "region" {
  value = local.vars.region
}

output "eks_role_arn" {
  value = module.iam_eks_role.iam_role_arn
}

output "ecr_repository_url" {
  value = data.aws_ecr_repository.ecr_repository.repository_url
}

output "ecr_registry_id" {
  value = data.aws_ecr_repository.ecr_repository.registry_id
}

output "cluster_name" {
  value = "${var.env}-nuon"
}

output "cluster_gh_role_arn" {
  value = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.github_action_role_arn)
}
