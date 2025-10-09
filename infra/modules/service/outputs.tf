output "release" {
  description = "Outputs required for releasing a service, such as pushing a helm chart or ECR image."

  value = {
    gh_role_arn        = aws_iam_role.github_actions.arn
    region             = local.vars.region
    eks_role_arn       = module.iam_eks_role.iam_role_arn
    ecr_repository_url = data.aws_ecr_repository.ecr_repository.repository_url
    ecr_registry_id    = data.aws_ecr_repository.ecr_repository.registry_id
  }
}

output "deploy" {
  description = "Outputs required for deploying a service, such as deploying a helm chart to a cluster."

  value = {
    region              = local.vars.region
    ecr_repository_url  = data.aws_ecr_repository.ecr_repository.repository_url
    ecr_registry_id     = data.aws_ecr_repository.ecr_repository.registry_id
    cluster_name        = "${var.env}-nuon"
    cluster_gh_role_arn = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.github_action_role_arn)
  }
}

output "tags" {
  value = local.tags
}

// NOTE(jm): legacy outputs, until we update our services to read these
output "gh_role_arn" {
  value = aws_iam_role.github_actions.arn
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
