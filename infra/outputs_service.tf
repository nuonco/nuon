output "installations_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.installations.name)
}

output "installations_k8s_role_arn" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.installations.k8s_access_role_arn)
}

output "orgs_k8s_role_arn" {
  # NOTE: you need to update `infra-eks` to add your service into the auth map
  value = nonsensitive(data.tfe_outputs.infra-eks-orgs.values.auth_map_additional_role_arns["eks-workers-installs"])
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

output "orgs_instance_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.instances_access)
}
