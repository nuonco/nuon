output "deployments_bucket_name" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets.deployments.name)
}

output "orgs_deployments_role_template" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates.deployments_access)
}

output "waypoint_server" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.waypoint)
}
