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
  value = data.aws_ecr_repository.api-gateway.repository_url
}

output "ecr_registry_id" {
  value = data.aws_ecr_repository.api-gateway.registry_id
}

output "certificate_arn" {
  value = module.certificate.acm_certificate_arn
}

output "certificate_domains" {
  value = module.certificate.distinct_domain_names
}

output "cluster_name" {
  value = "${var.env}-nuon"
}

output "cluster_gh_role_arn" {
  value = nonsensitive(data.tfe_outputs.infra-eks-nuon.values.github_action_role_arn)
}
