output "deployments_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.deployments.name)
}

output "orgs_deployments_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.deployments_access)
}

output "orgs_installations_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.installations_access)
}

output "installations_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.installations.name)
}

output "orgs_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.orgs.name)
}

output "orgs_orgs_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.orgs_access)
}

# the following values are for interacting with the orgs k8s cluster, and while they currently are not being used should
# be migrated too when possible.
output "orgs_k8s_role_arn" {
  # NOTE: you need to update `infra-eks` to add your service into the auth map
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.access_role_arns["eks-orgs-api"])
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

# the following are for accessing waypoint
output "waypoint_server_root_domain" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.waypoint.root_domain)
}

output "waypoint_token_secret_namespace" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.waypoint.token_secret_namespace)
}

output "waypoint_token_secret_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.waypoint.token_secret_template)
}
