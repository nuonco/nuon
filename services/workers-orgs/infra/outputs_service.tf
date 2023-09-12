# access for talking to the orgs cluster
output "orgs_k8s_role_arn" {
  # NOTE: you need to update `infra-eks` to add your service into the auth map
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.access_role_arns["eks-workers-orgs"])
}

output "orgs_k8s_cluster_id" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.cluster_id)
}

output "orgs_k8s_ca_data" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.ca_data)
}

output "orgs_k8s_public_endpoint" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.public_endpoint)
}

output "orgs_iam_oidc_provider_url" {
  description = "OIDC provider url (without leading https) in the orgs cluster"
  value       = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.oidc_provider_url)
}

output "orgs_iam_oidc_provider_arn" {
  description = "OIDC provider ARN in the orgs cluster"
  value       = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.oidc_provider_arn)
}

# configuration for orgs resources
output "orgs_ecr_registry_arn" {
  description = "Base ECR repo arn for orgs."
  value       = nonsensitive(data.tfe_outputs.infra-orgs.values.ecr.registry_arn)
}

output "org_installations_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.installations.name)
}

output "org_orgs_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.orgs.name)
}

output "org_deployments_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.deployments.name)
}

output "org_secrets_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.secrets.name)
}

# support role for accessing org IAM roles
output "support_iam_role_arn" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.iam_roles.support.arn)
}

output "orgs_account_root_arn" {
  value = "arn:aws:iam::${local.accounts["orgs-${var.env}"].id}:root"
}

// the following outputs are from resources managed locally
output "workers_iam_role_arn_prefix" {
  description = "iam role prefix for the worker service roles that should assume org iam roles"

  value = "arn:aws:iam::${local.accounts[var.env].id}:role/eks/eks-*"
}

output "orgs_account_iam_access_role_arn" {
  description = "IAM role for managing IAM resources in the orgs account"
  value       = module.orgs_account_iam_access_role.iam_role_arn
}

output "orgs_account_bucket_access_role_arn" {
  description = "IAM role for accessing the orgs bucket in the orgs account"
  value       = module.orgs_account_bucket_access_role.iam_role_arn
}

output "waypoint_server_root_domain" {
  description = "root domain for waypoint server"
  value       = "orgs-${var.env}.nuon.co"
}

output "waypoint_bootstrap_token_namespace" {
  description = "root domain for waypoint server"
  value       = "default"
}

output "orgs_account_kms_access_role_arn" {
  description = "IAM role for managing KMS resources in the orgs account"
  value       = module.orgs_account_kms_access_role.iam_role_arn
}

output "orgs_account" {
  description = "orgs account details"
  value       = nonsensitive(data.tfe_outputs.infra-orgs.values.account)
}

output "sandbox" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.sandbox)
}
