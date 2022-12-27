output "deployments_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.deployments.name)
}

output "orgs_deployments_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.deployments_access)
}

# the following values are for interacting with the orgs k8s cluster, and while they currently are not being used should
# be migrated too when possible.
output "orgs_k8s_role_arn" {
  # NOTE: you need to update `infra-eks` to add your service into the auth map
  value = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.auth_map_additional_role_arns["eks-workers-deployments"])
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
