# access installations
output "installations_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.installations.name)
}

output "installations_bucket_region" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.installations.region)
}

# this role is added to the sandbox, and grants Nuon access to do things in it
output "installations_k8s_role_arn" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.iam_roles.install_k8s_access)
}

# the following values are for interacting with the orgs k8s cluster
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

# the following values are for using IAM roles related to an org
output "orgs_instance_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.instances_access)
}

output "orgs_installations_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.installations_access)
}

output "orgs_installer_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.installer)
}
