output "buckets" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets)
}

output "org_iam_role_name_templates" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates)
}

# the following values are for interacting with the orgs k8s cluster, and while they currently are not being used should
# be migrated too when possible.
output "orgs_k8s" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s)
}

# the following values are for the plan child workflow
output "orgs_ecr" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.ecr)
}

# the following are for accessing waypoint
output "waypoint_server" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.waypoint)
}
