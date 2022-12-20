output "installations_k8s_role_arn" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.install_k8s_role_arn)
}

output "orgs_k8s_role_arn" {
  # NOTE: you need to update `infra-eks` to add your service into the auth map
  value = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.auth_map_additional_role_arns["eks-workers-orgs"])
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

output "orgs_account_iam_access_role_arn" {
  description = "IAM role for managing IAM resources in the orgs account"
  value       = module.orgs_account_iam_access_role.iam_role_arn
}

output "orgs_iam_oidc_provider_url" {
  description = "OIDC provider url (without leading https) in the orgs cluster"
  value       = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.oidc_provider)
}

output "orgs_iam_oidc_provider_arn" {
  description = "OIDC provider ARN in the orgs cluster"
  value       = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.oidc_provider_arn)
}

output "orgs_ecr_registry_arn" {
  description = "Base ECR repo arn for orgs."

  value = "arn:aws:ecr:${local.vars.region}:${local.accounts["orgs-${var.env}"].id}:repository"
}

output "org_installations_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets["installations"])
}

output "org_deployments_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets["deployments"])
}

output "workers_iam_role_arn_prefix" {
  description = "iam role prefix for the worker service roles that should assume org iam roles"

  value = "arn:aws:iam::${local.accounts[var.env].id}:role/eks/eks-workers-*"
}
