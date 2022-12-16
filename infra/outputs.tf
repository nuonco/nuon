output "gh_role_arn" {
  value = module.github_actions.iam_role_arn
}

output "region" {
  value = local.vars.region
}

output "eks_role_arn" {
  value = module.iam_eks_role.iam_role_arn
}

output "nuon_charts" {
  value = local.helm_bucket_url
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

output "installations_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.bucket_name)
}

output "installations_k8s_role_arn" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.install_k8s_role_arn)
}

output "orgs_k8s_role_arn" {
  # NOTE: you need to update `infra-eks` to add your service into the auth map
  value = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.auth_map_additional_role_arns["eks-workers-installs"])
}

output "orgs_k8s_cluster_id" {
  value = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_id)
}

output "orgs_k8s_ca_data" {
  value = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_certificate_authority_data)
}

output "orgs_k8s_public_endpoint" {
  value = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.cluster_endpoint)
}
